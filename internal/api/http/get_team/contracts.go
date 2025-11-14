package get_team

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type TeamService interface {
	GetTeamWithMembers(ctx context.Context, teamName string) (entity.Team, error)
}
