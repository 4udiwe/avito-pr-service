package get_prs

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type PRService interface {
	GetAllPRs(ctx context.Context, page, pageSize int) ([]entity.PullRequest, int, error)
}
