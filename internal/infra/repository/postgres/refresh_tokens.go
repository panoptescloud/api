package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panoptescloud/api/internal/infra/repository/postgres/db"
)

type RefreshTokensRepository struct {
	p *pgxpool.Pool
}

type token interface {
	ID() uuid.UUID
	UserID() uuid.UUID
	Value() string
	IssuedAt() time.Time
	ExpiresAt() time.Time
}

func (rtr *RefreshTokensRepository) Save(t token) error {
	queries := db.New(rtr.p)

	id := t.ID()
	userID := t.UserID()

	err := queries.InsertRefreshToken(context.TODO(), db.InsertRefreshTokenParams{
		ID: pgtype.UUID{
			Bytes: [16]byte(id[:]),
			Valid: true,
		},
		Value: t.Value(),
		IssuedAt: pgtype.Timestamptz{
			Time:  t.IssuedAt().UTC(), // always store in UTC
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  t.ExpiresAt().UTC(), // always store in UTC
			Valid: true,
		},
		UserID: pgtype.UUID{
			Bytes: [16]byte(userID[:]),
			Valid: true,
		},
	})

	return err
}

func NewRefreshTokensRepository(pool *pgxpool.Pool) *RefreshTokensRepository {
	return &RefreshTokensRepository{
		p: pool,
	}
}
