package user

import (
	"context"
	"errors"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	"github.com/4udiwe/avito-pr-service/pkg/transactor"
	"github.com/sirupsen/logrus"
)

type Service struct {
	userRepo  UserRepo
	PRRepo    PullReqeustRepo
	txManager transactor.Transactor
}

func New(userRepo UserRepo, PRRepo PullReqeustRepo, txManager transactor.Transactor) *Service {
	return &Service{
		userRepo:  userRepo,
		PRRepo:    PRRepo,
		txManager: txManager,
	}
}

func (s *Service) SetUserStatus(ctx context.Context, userID string, isActive bool) (entity.User, error) {
	logrus.Infof("UserService.SetUserStatus: setting user %s active status to %v", userID, isActive)

	err := s.userRepo.SetActiveStatus(ctx, userID, isActive)

	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			logrus.Warnf("UserService.SetUserStatus: user %s not found", userID)
			return entity.User{}, ErrUserNotFound
		}
		logrus.Errorf("UserService.SetUserStatus: failed to set active status for user %s: %v", userID, err)
		return entity.User{}, ErrCannotSetUserStatus
	}

	user, err := s.userRepo.GetByID(ctx, userID)

	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			logrus.Warnf("UserService.SetUserStatus: user %s not found after status update", userID)
			return entity.User{}, ErrUserNotFound
		}
		logrus.Errorf("UserService.SetUserStatus: failed to get user %s after status update: %v", userID, err)
		return entity.User{}, ErrCannotSetUserStatus
	}

	logrus.Infof("UserService.SetUserStatus: user %s active status set to %v", userID, isActive)
	return user, nil
}

func (s *Service) GetUserReviews(ctx context.Context, userID string) ([]entity.PullRequest, error) {
	logrus.Infof("UserService.GetUserReviews: fetching prs for user %s", userID)

	var prs []entity.PullRequest

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		// Check if user exists
		_, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			logrus.Warnf("UserService.GetUserReviews: user %s not found", userID)
			return err
		}

		// Fetch PRs assigned to the user as a reviewer
		prs, err = s.PRRepo.ListByReviewer(ctx, userID)
		return err
	})

	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		logrus.Errorf("UserService.GetUserReviews: failed to list PRs for reviewer %s: %v", userID, err)
		return nil, ErrCannotGetUserReviews
	}

	logrus.Infof("UserService.GetUserReviews: fetched %d PRs for user %s", len(prs), userID)
	return prs, nil
}
