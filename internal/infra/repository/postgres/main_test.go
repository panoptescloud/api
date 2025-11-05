package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panoptescloud/api/tests/db/postgrestest"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	container, basePool := postgrestest.SetupTestDB()

	pool = basePool
	defer pool.Close()

	exitCode := m.Run()

	if err := container.Terminate(ctx); err != nil {
		panic(err)
	}

	os.Exit(exitCode)
}

func newUuidV7(t *testing.T) uuid.UUID {
	id, err := uuid.NewUUID()

	if err != nil {
		t.Logf("failed to create uuid: %s", err.Error())
		t.FailNow()
	}

	return id
}

func createUserWithGithubUser(pool *pgxpool.Pool, id uuid.UUID, name string, email string, nodeID string) error {
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
