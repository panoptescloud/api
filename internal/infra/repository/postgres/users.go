package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panoptescloud/api/internal/domain"
	"github.com/panoptescloud/api/internal/domain/users"
	"github.com/panoptescloud/api/internal/infra/repository/postgres/db"
)

type UsersRepository struct {
	p *pgxpool.Pool
}

func (u *UsersRepository) ByGithubNodeId(id string) (*users.User, error) {
	queries := db.New(u.p)

	dbUser, err := queries.GetUserByGithubUserNodeID(context.TODO(), id)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	user, err := users.HydrateUser(
		dbUser.ID.String(),
		dbUser.Name,
		dbUser.Email,
		dbUser.NodeID,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UsersRepository) Save(user *users.User) error {
	queries := db.New(u.p)

	err := queries.UpsertUser(context.TODO(), db.UpsertUserParams{
		ID: pgtype.UUID{
			Bytes: [16]byte(user.ID().Bytes()),
			Valid: true,
		},
		Email: user.Email().String(),
		Name:  user.Name().String(),
	})

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_unique" {
				return domain.ErrEmailAlreadyInUse{}
			}
		}

		return err
	}

	err = queries.UpsertGithubUser(context.TODO(), db.UpsertGithubUserParams{
		UserID: pgtype.UUID{
			Bytes: [16]byte(user.ID().Bytes()),
			Valid: true,
		},
		NodeID: user.GithubIdentity().NodeID().String(),
	})

	return err
}

func NewUsersRepository(p *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{
		p: p,
	}
}
