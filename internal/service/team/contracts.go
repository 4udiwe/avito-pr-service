package team

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/google/uuid"
)

type UserRepo interface {
	Create(ctx context.Context, ID, name string, teamID uuid.UUID, isActive bool) (entity.User, error)
	GetByTeamID(ctx context.Context, teamID uuid.UUID) ([]entity.User, error)
}

type TeamRepo interface {
	Create(ctx context.Context, name string) (entity.Team, error)
	GetByName(ctx context.Context, name string) (entity.Team, error)
}
