package dto

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	Token     HashedValue
	IssuedAt  time.Time
	ExpiresAt time.Time
	UserID    uuid.UUID
}
