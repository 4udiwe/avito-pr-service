package get_stats

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type StatsService interface {
	GetStats(ctx context.Context) (*entity.Stats, error)
}
