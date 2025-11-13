package repo_user

import (
	"time"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/google/uuid"
)

type RowUser struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	TeamID    uuid.UUID `db:"team_id"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
}

func (ru *RowUser) ToEntity() entity.User {
	return entity.User{
		ID:        ru.ID,
		Name:      ru.Name,
		TeamID:    ru.TeamID,
		IsActive:  ru.IsActive,
		CreatedAt: ru.CreatedAt,
	}
}
