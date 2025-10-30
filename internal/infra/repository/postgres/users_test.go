package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panoptescloud/api/internal/infra/repository/postgres"
	"github.com/panoptescloud/api/tests/db/postgrestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func Test_UsersRepository_ByGithubNodeId_empty_DB(t *testing.T) {
	t.Parallel()

	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewUsersRepository(pool)

	found, err := repo.ByGithubNodeId("blah")

	assert.Nil(t, found)
	assert.Nil(t, err)
}

func Test_UsersRepository_ByGithubNodeId_finds_one(t *testing.T) {
	t.Parallel()

	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewUsersRepository(pool)

	id := newUuidV7(t)
	name := "josephus miller"
	email := "josephus@starhelix.org.ceres"
	nodeID := "some_node_id"
	err := createUserWithGithubUser(
		pool,
		id,
		name,
		email,
		nodeID,
	)

	require.Nil(t, err)

	found, err := repo.ByGithubNodeId("some_node_id")

	require.Nil(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, id.String(), found.ID().String())
	assert.Equal(t, name, found.Name().String())
	assert.Equal(t, email, found.Email().String())
	assert.Equal(t, nodeID, found.GithubIdentity().NodeID().String())
}

