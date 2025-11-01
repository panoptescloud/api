package postgrestest

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func SetupTestDB() (*tcpostgres.PostgresContainer, *pgxpool.Pool) {
	ctx := context.Background()

	dbName := "panoptes"
	dbUser := "panoptes"
	dbPassword := "iamtest"

	container, err := tcpostgres.Run(ctx,
		// TODO: pass this as an env var
		"timescale/timescaledb-ha:pg17",
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(dbUser),
		tcpostgres.WithPassword(dbPassword),
		tcpostgres.BasicWaitStrategies(),
	)

	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}

	dbURL, err := container.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	return container, pool
}

func createIsolatedSchema(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()

	ctx := context.Background()
	schema := fmt.Sprintf("test_%s", t.Name())

	// Replace invalid chars (like slashes from subtests)
	schema = strings.ReplaceAll(strings.ToLower(schema), "/", "_")

	_, err := pool.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, schema))
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	if err := runMigrationsForSchema(ctx, pool, schema); err != nil {
		t.Logf("failed to drop schema %s: %v", schema, err)
	}

	t.Cleanup(func() {
		_, err := pool.Exec(ctx, fmt.Sprintf(`DROP SCHEMA %s CASCADE;`, schema))
		if err != nil {
			t.Logf("failed to drop schema %s: %v", schema, err)
		}
	})

	return schema
}

func poolForSchema(ctx context.Context, basePool *pgxpool.Pool, schema string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(basePool.Config().ConnString())
	if err != nil {
		return nil, err
	}

	cfg.ConnConfig.RuntimeParams["search_path"] = schema

	return pgxpool.NewWithConfig(ctx, cfg)
}

func runMigrationsForSchema(ctx context.Context, pool *pgxpool.Pool, schema string) error {
	// _, err := pool.Exec(ctx, fmt.Sprintf("SET search_path TO %s;", schema))
	// if err != nil {
	// 	return err
	// }
	
	dir := "/app/etc/postgres/migrations"
	files := []string{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".sql" {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("reading migration dir: %w", err)
	}

	sort.Strings(files) // Ensure deterministic order

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading %s: %w", file, err)
		}

		_, err = pool.Exec(ctx, string(sqlBytes))
		if err != nil {
			return fmt.Errorf("executing %s: %w", file, err)
		}
	}

	return nil
}

func PrepareDBForTest(t *testing.T, initialPool *pgxpool.Pool, runMigrations bool) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	schema := createIsolatedSchema(t, initialPool)
	schemaPool, err := poolForSchema(ctx, initialPool, schema)

	if err != nil {
		log.Fatalf("failed to create schema pool(%s): %v", t.Name(), err)
	}

	if runMigrations {
		if err := runMigrationsForSchema(ctx, schemaPool, schema); err != nil {
			log.Fatalf("failed to run migrations for test: %s", t.Name())
		}
	}

	return schemaPool
}