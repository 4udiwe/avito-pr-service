package user

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type UserRepo interface {
	GetByID(ctx context.Context, ID string) (entity.User, error)
	SetActiveStatus(ctx context.Context, userID string, isActive bool) error
}

type PullReqeustRepo interface {
	ListByReviewer(ctx context.Context, reviewerID string) ([]entity.PullRequest, error)
}
