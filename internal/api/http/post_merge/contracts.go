package post_merge

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type PRService interface {
	MergePR(ctx context.Context, prID string) (entity.PullRequest, error)
}
