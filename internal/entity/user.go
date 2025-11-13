package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        string
	Name      string
	IsActive  bool
	TeamID    uuid.UUID
	CreatedAt time.Time
}
