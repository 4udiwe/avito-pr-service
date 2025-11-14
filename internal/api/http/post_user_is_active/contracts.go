package post_user_is_active

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type UserService interface {
	SetUserStatus(ctx context.Context, userID string, isActive bool) (entity.User, error)
}
