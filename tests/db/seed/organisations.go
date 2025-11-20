package seed

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Organisation(pool *pgxpool.Pool, id uuid.UUID, name string, memberID uuid.UUID, role string) error {
	_, err := pool.Exec(context.Background(), `
		INSERT INTO organisations (id, name) VALUES ($1, $2);
	`, id, name)

	if err != nil {
		return err
	}
	_, err = pool.Exec(context.Background(), `
		INSERT INTO organisation_members (organisation_id, member_id, role) VALUES ($1, $2, $3);
	`, id, memberID, role)

	return err
}
