package pr

import (
	"context"
	"errors"
	"time"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	"github.com/4udiwe/avito-pr-service/pkg/transactor"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

const ReviewerCount = 2

type Service struct {
	PRRepo    PRRepo
	UserRepo  UserRepo
	txManager transactor.Transactor
}

func New(prRepo PRRepo, userRepo UserRepo, txManager transactor.Transactor) *Service {
	return &Service{
		PRRepo:    prRepo,
		UserRepo:  userRepo,
		txManager: txManager,
	}
}

func (s *Service) CreatePR(ctx context.Context, pullRequestID, title, authorID string) (entity.PullRequest, error) {
	logrus.Infof("PRService.CreatePR: creating PR with title %s", title)

	var pullRequest entity.PullRequest

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Create the PR
		pr, err := s.PRRepo.Create(ctx, pullRequestID, title, authorID)
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
		pullRequest = pr

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
		err = s.PRRepo.AssignReviewers(ctx, pullRequestID, reviewerIDs)
		if err != nil {
			return err
		}

		pullRequest.Reviewers = reviewerIDs
		return nil
	})

	if err != nil {
		if errors.Is(err, repository.ErrReviewerAlreadyAssigned) {
			logrus.Warnf("PRService.CreatePR: reviewer already assigned to PR %s", title)
			return entity.PullRequest{}, ErrReviewerAlreadyAssigned
		}
		if errors.Is(err, repository.ErrReviewerNotFound) {
			logrus.Warnf("PRService.CreatePR: reviewer not found for PR %s", title)
			return entity.PullRequest{}, ErrReviewerNotFound
		}
		return entity.PullRequest{}, ErrCannotCreatePR
	}

	logrus.Infof("PRService.CreatePR: successfully created PR %s", title)
	return pullRequest, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) (entity.PullRequest, error) {
	logrus.Infof("PRService.ReassignReviewer: reassigning reviewer for PR %s", prID)

	var pullRequest entity.PullRequest
	var reviewers []entity.PRReviewer

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Check if PR exists
		pr, err := s.PRRepo.GetByID(ctx, prID)
		if err != nil {
			if errors.Is(err, repository.ErrPRNotFound) {
				logrus.Warnf("PRService.ReassignReviewer: PR with ID %s not found", prID)
				return ErrPRNotFound
			}
			logrus.Errorf("PRService.ReassignReviewer: failed to get PR with ID %s: %v", prID, err)
			return err
		}

		pullRequest = pr

		// Check status of the PR
		pullRequest.Status, err = s.PRRepo.GetStatusByStatusID(ctx, pullRequest.Status.ID)
		if err != nil {
			if errors.Is(err, repository.ErrStatusNotFound) {
				logrus.Warnf("PRService.ReassignReviewer: status with ID %d not found for PR %s", pullRequest.Status.ID, prID)
				return ErrStatusNotFound
			}
			logrus.Errorf("PRService.ReassignReviewer: failed to get status with ID %d for PR %s: %v", pullRequest.Status.ID, prID, err)
			return err
		}

		if pullRequest.Status.Name == entity.StatusMERGED {
			logrus.Warnf("PRService.ReassignReviewer: cannot reassign reviewer for merged PR %s", prID)
			return ErrCannotReassignReviewerForMergedPR
		}

		// Reassign reviewer
		err = s.PRRepo.ReassignReviewer(ctx, prID, oldReviewerID, newReviewerID)
		if err != nil {
			return err
		}

		// Get updated list of reviewers
		reviewers, err = s.PRRepo.GetReviewersByPR(ctx, prID)
		return err
	})

	if err != nil {
		if errors.Is(err, repository.ErrReviewerNotFound) {
			logrus.Warnf("PRService.ReassignReviewer: reviewer not found for PR %s", prID)
			return entity.PullRequest{}, ErrReviewerNotFound
		}
	}

	pullRequest.Reviewers = lo.Map(reviewers, func(r entity.PRReviewer, _ int) string { return r.ID })

	return pullRequest, nil
}

func (s *Service) MergePR(ctx context.Context, prID string) (entity.PullRequest, error) {
	logrus.Infof("PRService.MergePR: merging PR %s", prID)

	var pullRequest entity.PullRequest

	// Get ID of MERGED status
	statuses, err := s.PRRepo.GetPRStatuses(ctx)
	if err != nil {
		logrus.Errorf("PRService.MergePR: failed to get PR statuses for PR %s: %v", prID, err)
		return entity.PullRequest{}, err
	}

	var mergedStatusID int
	for _, status := range statuses {
		if status.Name == entity.StatusMERGED {
			mergedStatusID = status.ID
			break
		}
	}

	err = s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Update PR status to MERGED
		err = s.PRRepo.UpdateStatus(ctx, prID, mergedStatusID, time.Now())
		if err != nil {
			if errors.Is(err, repository.ErrPRNotFound) {
				logrus.Warnf("PRService.MergePR: PR with ID %s not found", prID)
				return ErrPRNotFound
			}
			logrus.Errorf("PRService.MergePR: failed to update status for PR %s: %v", prID, err)
			return err
		}

		// Get updated PR with reviewers
		pullRequest, err = s.PRRepo.GetByID(ctx, prID)
		return err
	})

	if err != nil {
		logrus.Errorf("PRService.MergePR: failed to merge PR %s: %v", prID, err)
		return entity.PullRequest{}, ErrCannotMergePR
	}

	logrus.Infof("PRService.MergePR: successfully merged PR %s", prID)
	return pullRequest, nil
}
