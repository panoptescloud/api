package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panoptescloud/api/internal/infra/repository/postgres/db"
	"github.com/panoptescloud/api/pkg/dto"
)

type RefreshTokensRepository struct {
	p *pgxpool.Pool
}

func (rtr *RefreshTokensRepository) Save(t dto.RefreshToken) error {
	queries := db.New(rtr.p)

	id := t.ID
	userID := t.UserID

	err := queries.InsertRefreshToken(context.TODO(), db.InsertRefreshTokenParams{
		ID: pgtype.UUID{
			Bytes: [16]byte(id[:]),
			Valid: true,
		},
		Value: t.Token.Value,
		IssuedAt: pgtype.Timestamptz{
			Time:  t.IssuedAt.UTC(), // always store in UTC
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  t.ExpiresAt.UTC(), // always store in UTC
			Valid: true,
		},
		UserID: pgtype.UUID{
			Bytes: [16]byte(userID[:]),
			Valid: true,
		},
	})

	return err
}

func (rtr *RefreshTokensRepository) ByToken(token dto.HashedValue) (dto.RefreshToken, error) {
	queries := db.New(rtr.p)

	t, err := queries.GetRefreshTokenByValue(context.TODO(), token.Value)

	if err != nil {
		return dto.RefreshToken{}, err
	}

	return dto.RefreshToken{
		ID: t.ID.Bytes,
		UserID: t.UserID.Bytes,
		Token: dto.HashedValue{
			Value: t.Value,
		},
		IssuedAt: t.IssuedAt.Time,
		ExpiresAt: t.ExpiresAt.Time,
	}, nil
}

func (rtr *RefreshTokensRepository) Delete(id uuid.UUID) error {
	queries := db.New(rtr.p)

	return queries.Delete(context.TODO(), pgtype.UUID{
		Bytes: [16]byte(id[:]),
		Valid: true,
	})
}

func NewRefreshTokensRepository(pool *pgxpool.Pool) *RefreshTokensRepository {
	return &RefreshTokensRepository{
		p: pool,
	}
}
