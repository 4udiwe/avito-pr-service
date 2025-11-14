package post_team

import (
	"context"

	"github.com/4udiwe/avito-pr-service/internal/entity"
)

type TeamService interface {
	CreateTeamWithUsers(ctx context.Context, teamName string, users []entity.User) (entity.Team, error)
}
