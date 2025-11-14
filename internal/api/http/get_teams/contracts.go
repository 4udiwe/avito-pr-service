package get_teams

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type TeamService interface {
	GetAllTeams(ctx context.Context, page, pageSize int) ([]entity.Team, int, error)
}
