package post_pr

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type PRService interface {
	CreatePR(ctx context.Context, pullRequestID, title, authorID string) (entity.PullRequest, error)
}
