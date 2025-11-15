package post_deactivate_team

import (
	"context"
)

type TeamService interface {
	DeactivateTeamAndReassignPRs(ctx context.Context, teamName string) error
}
