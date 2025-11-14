package post_assign

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type PRService interface {
	AssignReviewer(ctx context.Context, prID, newReviewerID string) (entity.PullRequest, error)
}
