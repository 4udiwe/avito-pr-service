package repo_team

import (
	"time"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/google/uuid"
)

type RowTeam struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

func (rt *RowTeam) ToEntity() entity.Team {
	return entity.Team{
		ID:        rt.ID,
		Name:      rt.Name,
		CreatedAt: rt.CreatedAt,
	}
}
