package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/organisations/domain"
	"github.com/panoptescloud/api/internal/organisations/infra/postgres/db"
)

type APIKeysRepository struct {
	p *pgxpool.Pool
}

func (u *APIKeysRepository) ByToken(token dto.HashedValue) (*domain.APIKey, error) {
	queries := db.New(u.p)

	key, err := queries.APIKeyByToken(context.TODO(), token.Value)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return domain.HydrateAPIKey(key.ID.String(), key.Name, key.OrganisationID.String(), key.Token)
}

func (u *APIKeysRepository) Save(key *domain.APIKey) error {
	queries := db.New(u.p)

	err := queries.UpsertAPIKey(context.TODO(), db.UpsertAPIKeyParams{
		ID: pgtype.UUID{
			Bytes: [16]byte(key.ID().Bytes()),
			Valid: true,
		},
		OrganisationID: pgtype.UUID{
			Bytes: [16]byte(key.OrganisationID().Bytes()),
			Valid: true,
		},
		Token: key.Token().Value,
		Name:  key.Name().String(),
	})

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "organisation_api_key_name_unique" {
				return domain.ErrAPIKeyNameAlreadyInUse{}
			}

			if pgErr.Code == "23505" && pgErr.ConstraintName == "organisation_api_key_token_unique" {
				// Ensure this is never surfaced to a user...should be handled
				// explicitly by callers
				return domain.ErrTokenConflict{}
			}
		}

		return err
	}

	return nil
}

func NewAPIKeysRepository(p *pgxpool.Pool) *APIKeysRepository {
	return &APIKeysRepository{
		p: p,
	}
}
