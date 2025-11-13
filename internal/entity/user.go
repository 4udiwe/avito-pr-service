package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	IsActive  bool      `db:"is_active"`
	TeamID    uuid.UUID `db:"team_id"`
	CreatedAt time.Time `db:"created_at"`
}
