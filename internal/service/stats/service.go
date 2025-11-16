package stats

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/sirupsen/logrus"
)

type Service struct {
	statsRepo StatsRepo
}

func New(statsRepo StatsRepo) *Service {
	return &Service{
		statsRepo: statsRepo,
	}
}

func (s *Service) GetStats(ctx context.Context) (*entity.Stats, error) {
	stats, err := s.statsRepo.GetStats(ctx)
	if err != nil {
		logrus.Errorf("Falied to collect stats: %v", err)
		return nil, ErrCannotCollectStats
	}
	return stats, nil
}
