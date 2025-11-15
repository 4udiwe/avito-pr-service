package pr

import (
	"context"
	"time"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/google/uuid"
)

type PRRepo interface {
	Create(ctx context.Context, ID, title, authorID, statusName string, needMoreReviewers bool) (entity.PullRequest, error)
	GetAll(ctx context.Context, limit int, offset int) (PRs []entity.PullRequest, total int, err error)
	AssignReviewers(ctx context.Context, prID string, reviewerIDs []string) error
	ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	GetByID(ctx context.Context, ID string) (entity.PullRequest, error)
	UpdateStatus(ctx context.Context, ID string, statusID int, mergedAt time.Time) error
	GetReviewersByPR(ctx context.Context, prID string) ([]entity.PRReviewer, error)
	GetPRStatuses(ctx context.Context) ([]entity.Status, error)
	GetStatusByStatusID(ctx context.Context, statusID int) (entity.Status, error)
	AssignReviewer(ctx context.Context, prID string, reviewerID string) error
	UpdateNeedMoreReviewers(ctx context.Context, ID string) error
}

type UserRepo interface {
	GetByID(ctx context.Context, ID string) (entity.User, error)
	GetRandomActiveTeammates(ctx context.Context, teamID uuid.UUID, limit int, excludeIDs ...string) ([]entity.User, error)
}
