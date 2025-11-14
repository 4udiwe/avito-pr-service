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
		candidates, err := s.UserRepo.GetRandomActiveTeammates(ctx, author.Team.ID, author.ID, ReviewerCount)
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

func (s *Service) GetAllPRs(ctx context.Context, page, pageSize int) ([]entity.PullRequest, int, error) {
	logrus.Info("PRService.GetAllPRs: fetching all PRs")

	limit := pageSize
	offset := (page - 1) * pageSize

	PRs, total, err := s.PRRepo.GetAll(ctx, limit, offset)
	if err != nil {
		logrus.Errorf("PRService.GetAllPRs: failed to fetch PRs %v", err)
		return nil, 0, ErrCannotFetchPRs
	}

	logrus.Infof("PRService.GetAllPRs: fetched %d PRs", len(PRs))
	return PRs, total, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (entity.PullRequest, string, error) {
	logrus.Infof("PRService.ReassignReviewer: reassigning reviewer for PR %s", prID)

	var pullRequest entity.PullRequest
	var reviewers []entity.PRReviewer
	var newReviewer entity.User

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
		if pullRequest.Status.Name == entity.StatusMERGED {
			logrus.Warnf("PRService.ReassignReviewer: cannot reassign reviewer for merged PR %s", prID)
			return ErrCannotReassignReviewerForMergedPR
		}

		// Get old reviewer with team
		oldReviewer, err := s.UserRepo.GetByID(ctx, oldReviewerID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				logrus.Warnf("PRService.ReassignReviewer: Old reviewer with ID %s not found", oldReviewerID)
			}
			return ErrReviewerNotFound
		}

		// Get random teammate (limit = 1) not the old reviewer
		reviewers, err := s.UserRepo.GetRandomActiveTeammates(ctx, oldReviewer.Team.ID, oldReviewerID, 1)
		if err != nil {
			return err
		}
		if len(reviewers) == 0 {
			return ErrNoMoreReviewersToReassign
		}
		newReviewer = reviewers[0]

		// Reassign reviewer
		return s.PRRepo.ReassignReviewer(ctx, prID, oldReviewerID, newReviewer.ID)
	})

	if err != nil {
		if errors.Is(err, repository.ErrReviewerNotFound) {
			logrus.Warnf("PRService.ReassignReviewer: reviewer not found for PR %s", prID)
			return entity.PullRequest{}, "", ErrReviewerNotFound
		}
	}

	// Get updated list of reviewers
	reviewers, err = s.PRRepo.GetReviewersByPR(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, "", ErrReviewerNotFound
	}

	pullRequest.Reviewers = lo.Map(reviewers, func(r entity.PRReviewer, _ int) string { return r.ID })

	return pullRequest, newReviewer.ID, nil
}

func (s *Service) MergePR(ctx context.Context, prID string) (entity.PullRequest, error) {
	logrus.Infof("PRService.MergePR: merging PR %s", prID)

	var pullRequest entity.PullRequest

	// Check if PR is already MERGED
	pullRequest, err := s.PRRepo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, repository.ErrPRNotFound) {
			logrus.Warnf("PRService.MergePR: PR with ID %s not found", prID)
			return entity.PullRequest{}, ErrPRNotFound
		}
		logrus.Errorf("PRService.MergePR: failed to PR %s: %v", prID, err)
		return entity.PullRequest{}, ErrCannotMergePR
	}

	if pullRequest.Status.Name == entity.StatusMERGED {
		return pullRequest, nil
	}

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

	// Update PR status to MERGED
	err = s.PRRepo.UpdateStatus(ctx, prID, mergedStatusID, time.Now())
	if err != nil {
		if errors.Is(err, repository.ErrPRNotFound) {
			logrus.Warnf("PRService.MergePR: PR with ID %s not found", prID)
			return entity.PullRequest{}, ErrPRNotFound
		}
		logrus.Errorf("PRService.MergePR: failed to update status for PR %s: %v", prID, err)
		return entity.PullRequest{}, ErrCannotCreatePR
	}

	// Get updated PR with reviewers
	pullRequest, err = s.PRRepo.GetByID(ctx, prID)
	if err != nil {
		logrus.Errorf("PRService.MergePR: failed to merge PR %s: %v", prID, err)
		return entity.PullRequest{}, ErrCannotMergePR
	}

	logrus.Infof("PRService.MergePR: successfully merged PR %s", prID)
	return pullRequest, nil
}

func (s *Service) AssignReviewer(ctx context.Context, prID, newReviewerID string) (entity.PullRequest, error) {
	logrus.Infof("PRService.AssignReviewer: assigning new reviewer %s for PR %s", newReviewerID, prID)

	var pullRequest entity.PullRequest

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Get PR
		pr, err := s.PRRepo.GetByID(ctx, prID)
		if err != nil {
			logrus.Infof("PRService.AssignReviewer: failed to get PR %s", prID)
			return err
		}

		// Check need_more_reviewers flag on PR
		if !pr.NeedMoreReviewers {
			return ErrPRAlreadyHas2Reviewers
		}

		// Assign new reviewer
		err = s.PRRepo.AssignReviewer(ctx, prID, newReviewerID)
		if err != nil {
			if errors.Is(err, repository.ErrReviewerNotFound) {
				logrus.Warnf("PRService.AssignReviewer: new reviewer %s not found", newReviewerID)
				return ErrReviewerNotFound
			}
			if errors.Is(err, repository.ErrReviewerAlreadyAssigned) {
				logrus.Warnf("PRService.AssignReviewer: new reviewer %s already assigned", newReviewerID)
				return ErrReviewerAlreadyAssigned
			}
			return err
		}

		// Check old amount of reviewers:
		// if was 1 (+ 1 new = 2) -> change need_more_reviewers to FALSE
		// if was 0 -> do nothing
		if len(pr.Reviewers) > 0 {
			err = s.PRRepo.UpdateNeedMoreReviewers(ctx, prID)
		}

		return err
	})

	if err != nil {
		logrus.Errorf("PRService.AssignReviewer: failed to assign new reviewer %s", newReviewerID)
		return entity.PullRequest{}, ErrCannotAssignReviewer
	}

	// Get updated list of reviewers
	reviewers, err := s.PRRepo.GetReviewersByPR(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, ErrReviewerNotFound
	}

	pullRequest.Reviewers = lo.Map(reviewers, func(r entity.PRReviewer, _ int) string { return r.ID })

	return pullRequest, nil
}
