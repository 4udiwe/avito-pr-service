package post_reassign

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type PRService interface {
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (entity.PullRequest, string, error)
}
