package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/panoptescloud/api/internal/infra/repository/postgres"
	"github.com/panoptescloud/api/tests/db/postgrestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tokenStub struct {
	id        uuid.UUID
	value     string
	issuedAt  time.Time
	expiresAt time.Time
	userID    uuid.UUID
}

func (ts tokenStub) ID() uuid.UUID {
	return ts.id
}

func (ts tokenStub) UserID() uuid.UUID {
	return ts.userID
}

func (ts tokenStub) Value() string {
	return ts.value
}

func (ts tokenStub) IssuedAt() time.Time {
	return ts.issuedAt
}

func (ts tokenStub) ExpiresAt() time.Time {
	return ts.expiresAt
}

func Test_RefreshTokensRepository_Save(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	userID := newUuidV7(t)
	err := createUserWithGithubUser(pool, userID, "josephus miller", "josephus@starhelix.org.ceres", "some_node_id")
	require.Nil(t, err)

	repo := postgres.NewRefreshTokensRepository(pool)

	token := tokenStub{
		id:        newUuidV7(t),
		userID:    userID,
		issuedAt:  time.Date(2025, time.November, 2, 12, 30, 0, 0, time.UTC),
		expiresAt: time.Date(2024, time.November, 3, 12, 30, 0, 0, time.UTC),
		value:     "somehashedstring",
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
		token.id,
	)

	err = row.Scan(&dbID, &dbUserID, &dbIssuedAt, &dbExpiresAt, &dbValue)
	require.NoError(t, err)

	assert.Equal(t, token.id, dbID)
	assert.Equal(t, token.userID, dbUserID)
	assert.WithinDuration(t, token.issuedAt, dbIssuedAt, time.Second)
	assert.WithinDuration(t, token.expiresAt, dbExpiresAt, time.Second)
	assert.Equal(t, token.value, dbValue)

}
