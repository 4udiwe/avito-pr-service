package stats

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type StatsRepo interface {
	GetStats(ctx context.Context) (*entity.Stats, error)
}
