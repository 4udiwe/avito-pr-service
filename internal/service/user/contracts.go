package user

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

//go:generate go tool mockgen -source=contracts.go -destination=mocks/mocks.go -package=mocks

type UserRepo interface {
	GetByID(ctx context.Context, ID string) (entity.User, error)
	SetActiveStatus(ctx context.Context, userID string, isActive bool) error
}

type PullReqeustRepo interface {
	ListByReviewer(ctx context.Context, reviewerID string) ([]entity.PullRequest, error)
}
