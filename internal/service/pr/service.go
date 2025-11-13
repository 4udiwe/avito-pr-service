package pr

import (
	"context"
	"errors"
	"time"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	"github.com/4udiwe/avito-pr-service/pkg/transactor"
	"github.com/sirupsen/logrus"
)

const (
	ReviewerCount = 2

	OpenStatus   = "OPEN"
	MergedStatus = "MERGED"
)

type Service struct {
	PrRepo    PrRepo
	UserRepo  UserRepo
	txManager transactor.Transactor
}

func New(prRepo PrRepo, userRepo UserRepo, txManager transactor.Transactor) *Service {
	return &Service{
		PrRepo:    prRepo,
		UserRepo:  userRepo,
		txManager: txManager,
	}
}

func (s *Service) CreatePR(ctx context.Context, pullRequestID, title, authorID string) (entity.PRWithReviewerIDs, error) {
	logrus.Infof("PRService.CreatePR: creating PR with title %s", title)

	var PRWithReviewerIDs entity.PRWithReviewerIDs

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Create the PR
		pr, err := s.PrRepo.Create(ctx, pullRequestID, title, authorID)
		if err != nil {
			if errors.Is(err, repository.ErrPRAlreadyExists) {
				logrus.Warnf("PRService.CreatePR: PR %s already exists", title)
				return err
			}
			if errors.Is(err, repository.ErrAuthorNotFound) {
				logrus.Warnf("PRService.CreatePR: author not found for PR %s", title)
				return err
			}
		}

		logrus.Infof("PRService.CreatePR: created PR %s with ID %s", pr.Title, pr.ID)
		PRWithReviewerIDs.PullRequest = pr

		// Get PR author
		author, err := s.UserRepo.GetByID(ctx, authorID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				logrus.Warnf("PRService.CreatePR: author with ID %s not found", authorID)
				return ErrAuthorNotFound
			}
			logrus.Errorf("PRService.CreatePR: failed to get author with ID %s: %v", authorID, err)
			return err
		}

		// Get 2 random author`s teammates
		candidates, err := s.UserRepo.GetRandomActiveTeammates(ctx, author.TeamID, author.ID, ReviewerCount)
		if err != nil {
			logrus.Errorf("PRService.CreatePR: failed to get teammates for author ID %s: %v", author.ID, err)
			return err
		}

		reviewerIDs := make([]string, len(candidates))
		for i, u := range candidates {
			reviewerIDs[i] = u.ID
		}

		// Assign reviewers
		err = s.PrRepo.AssignReviewers(ctx, pullRequestID, reviewerIDs)
		if err != nil {
			return err
		}

		PRWithReviewerIDs.ReviewersIDs = reviewerIDs
		return nil
	})

	if err != nil {
		if errors.Is(err, repository.ErrReviewerAlreadyAssigned) {
			logrus.Warnf("PRService.CreatePR: reviewer already assigned to PR %s", title)
			return entity.PRWithReviewerIDs{}, ErrReviewerAlreadyAssigned
		}
		if errors.Is(err, repository.ErrReviewerNotFound) {
			logrus.Warnf("PRService.CreatePR: reviewer not found for PR %s", title)
			return entity.PRWithReviewerIDs{}, ErrReviewerNotFound
		}
		return entity.PRWithReviewerIDs{}, ErrCannotCreatePR
	}

	logrus.Infof("PRService.CreatePR: successfully created PR %s", title)
	return PRWithReviewerIDs, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) (entity.PRWithReviewerIDs, error) {
	logrus.Infof("PRService.ReassignReviewer: reassigning reviewer for PR %s", prID)

	var prWithReviewerIDs entity.PRWithReviewerIDs
	var reviewers []entity.PRReviewer

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Check if PR exists
		pr, err := s.PrRepo.GetByID(ctx, prID)
		if err != nil {
			if errors.Is(err, repository.ErrPRNotFound) {
				logrus.Warnf("PRService.ReassignReviewer: PR with ID %s not found", prID)
				return ErrPRNotFound
			}
			logrus.Errorf("PRService.ReassignReviewer: failed to get PR with ID %s: %v", prID, err)
			return err
		}

		// Check status of the PR
		statuses, err := s.PrRepo.GetPRStatuses(ctx)
		if err != nil {
			logrus.Errorf("PRService.ReassignReviewer: failed to get status for PR %s: %v", prID, err)
			return err
		}

		var mergedStatusID int
		for _, status := range statuses {
			if status.Name == MergedStatus {
				mergedStatusID = status.ID
				break
			}
		}

		if pr.StatusID == mergedStatusID {
			logrus.Warnf("PRService.ReassignReviewer: cannot reassign reviewer for merged PR %s", prID)
			return ErrCannotReassignReviewerForMergedPR
		}

		// Reassign reviewer
		err = s.PrRepo.ReassignReviewer(ctx, prID, oldReviewerID, newReviewerID)
		if err != nil {
			return err
		}

		// Get updated list of reviewers
		reviewers, err = s.PrRepo.GetReviewersByPR(ctx, prID)
		return err
	})

	if err != nil {
		if errors.Is(err, repository.ErrReviewerNotFound) {
			logrus.Warnf("PRService.ReassignReviewer: reviewer not found for PR %s", prID)
			return entity.PRWithReviewerIDs{}, ErrReviewerNotFound
		}
	}

	prWithReviewerIDs.ReviewersIDs = make([]string, len(reviewers))
	for i, r := range reviewers {
		prWithReviewerIDs.ReviewersIDs[i] = r.ReviewerID
	}

	return prWithReviewerIDs, nil
}

func (s *Service) MergePR(ctx context.Context, prID string) (entity.PRWithReviewerIDs, error) {
	logrus.Infof("PRService.MergePR: merging PR %s", prID)

	var prWithReviewerIDs entity.PRWithReviewerIDs
	var reviewers []entity.PRReviewer

	// Get ID of MERGED status
	statuses, err := s.PrRepo.GetPRStatuses(ctx)
	if err != nil {
		logrus.Errorf("PRService.MergePR: failed to get PR statuses for PR %s: %v", prID, err)
		return entity.PRWithReviewerIDs{}, err
	}

	var mergedStatusID int
	for _, status := range statuses {
		if status.Name == MergedStatus {
			mergedStatusID = status.ID
			break
		}
	}

	err = s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Update PR status to MERGED
		err = s.PrRepo.UpdateStatus(ctx, prID, mergedStatusID, time.Now())
		if err != nil {
			if errors.Is(err, repository.ErrPRNotFound) {
				logrus.Warnf("PRService.MergePR: PR with ID %s not found", prID)
				return ErrPRNotFound
			}
			logrus.Errorf("PRService.MergePR: failed to update status for PR %s: %v", prID, err)
			return err
		}

		// Get updated PR
		pr, err := s.PrRepo.GetByID(ctx, prID)
		if err != nil {
			return err
		}
		prWithReviewerIDs.PullRequest = pr
		// Get reviewers
		reviewers, err = s.PrRepo.GetReviewersByPR(ctx, prID)
		if err != nil {
			return err
		}
		prWithReviewerIDs.ReviewersIDs = make([]string, len(reviewers))
		for i, r := range reviewers {
			prWithReviewerIDs.ReviewersIDs[i] = r.ReviewerID
		}
		return nil
	})

	if err != nil {
		logrus.Errorf("PRService.MergePR: failed to merge PR %s: %v", prID, err)
		return entity.PRWithReviewerIDs{}, ErrCannotMergePR
	}

	logrus.Infof("PRService.MergePR: successfully merged PR %s", prID)
	return prWithReviewerIDs, nil
}
