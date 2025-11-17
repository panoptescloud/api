package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/infra/repository/postgres"
	"github.com/panoptescloud/api/tests/db/postgrestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RefreshTokensRepository_Save(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	userID := newUuidV7(t)
	err := createUserWithGithubUser(pool, userID, "josephus miller", "josephus@starhelix.org.ceres", "some_node_id")
	require.Nil(t, err)

	repo := postgres.NewRefreshTokensRepository(pool)

	token := dto.RefreshToken{
		ID:        newUuidV7(t),
		UserID:    userID,
		IssuedAt:  time.Date(2025, time.November, 2, 12, 30, 0, 0, time.UTC),
		ExpiresAt: time.Date(2024, time.November, 3, 12, 30, 0, 0, time.UTC),
		Token: dto.HashedValue{
			Value: "somehashedstring",
		},
	}

	require.Nil(t, repo.Save(token))

	var (
		dbID        uuid.UUID
		dbUserID    uuid.UUID
		dbIssuedAt  time.Time
		dbExpiresAt time.Time
		dbValue     string
	)

	row := pool.QueryRow(
		context.Background(),
		`SELECT id, user_id, issued_at, expires_at, value FROM refresh_tokens WHERE id = $1`,
		token.ID,
	)

	err = row.Scan(&dbID, &dbUserID, &dbIssuedAt, &dbExpiresAt, &dbValue)
	require.NoError(t, err)

	assert.Equal(t, token.ID, dbID)
	assert.Equal(t, token.UserID, dbUserID)
	assert.WithinDuration(t, token.IssuedAt, dbIssuedAt, time.Second)
	assert.WithinDuration(t, token.ExpiresAt, dbExpiresAt, time.Second)
	assert.Equal(t, token.Token.Value, dbValue)

}

func Test_RefreshTokensRepository_Delete(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	userID := newUuidV7(t)
	err := createUserWithGithubUser(pool, userID, "josephus miller", "josephus@starhelix.org.ceres", "some_node_id")
	require.Nil(t, err)

	repo := postgres.NewRefreshTokensRepository(pool)

	token := dto.RefreshToken{
		ID:        newUuidV7(t),
		UserID:    userID,
		IssuedAt:  time.Date(2025, time.November, 2, 12, 30, 0, 0, time.UTC),
		ExpiresAt: time.Date(2024, time.November, 3, 12, 30, 0, 0, time.UTC),
		Token: dto.HashedValue{
			Value: "somehashedstring",
		},
	}

	require.Nil(t, repo.Save(token))

	var (
		dbID        uuid.UUID
		dbUserID    uuid.UUID
		dbIssuedAt  time.Time
		dbExpiresAt time.Time
		dbValue     string
	)

	row := pool.QueryRow(
		context.TODO(),
		`SELECT id, user_id, issued_at, expires_at, value FROM refresh_tokens WHERE id = $1`,
		token.ID,
	)

	err = row.Scan(&dbID, &dbUserID, &dbIssuedAt, &dbExpiresAt, &dbValue)
	require.NoError(t, err)

	assert.Equal(t, token.ID, dbID)
	assert.Equal(t, token.UserID, dbUserID)
	assert.WithinDuration(t, token.IssuedAt, dbIssuedAt, time.Second)
	assert.WithinDuration(t, token.ExpiresAt, dbExpiresAt, time.Second)
	assert.Equal(t, token.Token.Value, dbValue)

	require.Nil(t, repo.Delete(token.ID))

	var count int

	err = pool.QueryRow(
		context.TODO(),
		`SELECT COUNT(*) FROM refresh_tokens`,
	).Scan(&count)

	require.Nil(t, err)
	require.Equal(t, 0, count, "expected refresh_tokens table to be empty")

}

func Test_RefreshTokensRepository_ByToken(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	userID := newUuidV7(t)
	err := createUserWithGithubUser(pool, userID, "josephus miller", "josephus@starhelix.org.ceres", "some_node_id")
	require.Nil(t, err)

	repo := postgres.NewRefreshTokensRepository(pool)

	token := dto.RefreshToken{
		ID:        newUuidV7(t),
		UserID:    userID,
		IssuedAt:  time.Date(2025, time.November, 2, 12, 30, 0, 0, time.UTC),
		ExpiresAt: time.Date(2024, time.November, 3, 12, 30, 0, 0, time.UTC),
		Token: dto.HashedValue{
			Value: "somehashedstring",
		},
	}

	require.Nil(t, repo.Save(token))

	var (
		dbID        uuid.UUID
		dbUserID    uuid.UUID
		dbIssuedAt  time.Time
		dbExpiresAt time.Time
		dbValue     string
	)

	row := pool.QueryRow(
		context.TODO(),
		`SELECT id, user_id, issued_at, expires_at, value FROM refresh_tokens WHERE id = $1`,
		token.ID,
	)

	err = row.Scan(&dbID, &dbUserID, &dbIssuedAt, &dbExpiresAt, &dbValue)
	require.NoError(t, err)

	assert.Equal(t, token.ID, dbID)
	assert.Equal(t, token.UserID, dbUserID)
	assert.WithinDuration(t, token.IssuedAt, dbIssuedAt, time.Second)
	assert.WithinDuration(t, token.ExpiresAt, dbExpiresAt, time.Second)
	assert.Equal(t, token.Token.Value, dbValue)

	found, err := repo.ByToken(dto.HashedValue{
		Value: "somehashedstring",
	})
	require.Nil(t, err)

	assert.Equal(t, token.ID, found.ID)
	assert.Equal(t, token.UserID, found.UserID)
	assert.WithinDuration(t, token.IssuedAt, found.IssuedAt, time.Second)
	assert.WithinDuration(t, token.ExpiresAt, found.ExpiresAt, time.Second)
	assert.Equal(t, token.Token.Value, found.Token.Value)
}
