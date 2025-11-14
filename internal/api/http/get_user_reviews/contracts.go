package get_user_reviews

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type UserService interface {
	GetUserReviews(ctx context.Context, userID string) ([]entity.PullRequest, error)
}
