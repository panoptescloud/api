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

type Session struct {
	JWT                   string
	RefreshToken          RefreshToken
	CSRFToken             string
	UserID                string
	IssuedAt              time.Time
	JWTExpiresAt          time.Time
	RefreshTokenExpiresAt time.Time
}
