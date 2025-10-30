package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panoptescloud/api/internal/domain/users"
	"github.com/panoptescloud/api/internal/infra/repository/postgres/db"
)

type UsersRepository struct {
	p *pgxpool.Pool
}

func (u *UsersRepository) ByGithubNodeId(id string) (*users.User, error) {
	queries := db.New(u.p)

	dbUser, err := queries.GetUserByGithubUserNodeID(context.TODO(), id)

	fmt.Printf("%#v", err)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	user, err := users.HydrateUser(
		dbUser.ID.String(),
		dbUser.Name,
		dbUser.Email,
		&dbUser.NodeID,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}
func NewUsersRepository(p *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{
		p: p,
	}
}