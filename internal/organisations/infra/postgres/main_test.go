package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
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
		log.Errorf("failed to terminate container")
	}

	os.Exit(exitCode)
}
