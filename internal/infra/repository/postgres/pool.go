package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panoptescloud/api/pkg/config"
)

func NewPool(poolConfig *config.PostgresConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", 
		poolConfig.Username, 
		poolConfig.Password,
		poolConfig.Host,
		poolConfig.Port,
		poolConfig.DBName,
		string(poolConfig.SSLMode),
	)

	config, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		return nil, err
	}

	config.MaxConns = int32(poolConfig.MaxConnections)
	config.MinConns = int32(poolConfig.MinConnections)


	return pgxpool.NewWithConfig(context.Background(), config)
}

