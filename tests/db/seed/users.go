package seed

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func User(pool *pgxpool.Pool, id uuid.UUID, name string, email string, nodeID string) error {
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, email, name) VALUES ($1, $2, $3);
	`, id, email, name)

	if err != nil {
		return err
	}
	_, err = pool.Exec(context.Background(), `
		INSERT INTO github_users (user_id, node_id) VALUES ($1, $2);
	`, id, nodeID)

	return err
}
