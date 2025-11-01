package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panoptescloud/api/internal/domain"
	"github.com/panoptescloud/api/internal/domain/users"
	"github.com/panoptescloud/api/internal/infra/repository/postgres"
	"github.com/panoptescloud/api/tests/db/postgrestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getAllUsers(t *testing.T, pool *pgxpool.Pool) [][]any {
	rows, err := pool.Query(context.TODO(), "SELECT u.id, u.name, u.email, ghu.node_id FROM users u INNER JOIN github_users ghu ON u.id=ghu.user_id;")

	require.Nil(t, err)

	defer rows.Close()

	var allRows [][]any

	for rows.Next() {
		vals, err := rows.Values()
		require.Nil(t, err)
		allRows = append(allRows, vals)
	}
	require.Nil(t, rows.Err())

	return allRows
}

func assertUserEqualsRow(t *testing.T, user *users.User, dbRow []any) {
	dbID := uuid.UUID{}
	idBytes := dbRow[0].([16]uint8)
	copy(dbID[:], idBytes[:])

	require.Equal(t, user.ID().String(), dbID.String())
	require.Equal(t, user.Name().String(), dbRow[1].(string))
	require.Equal(t, user.Email().String(), dbRow[2].(string))
	require.Equal(t, user.GithubIdentity().NodeID().String(), dbRow[3].(string))
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

func Test_UsersRepository_ByGithubNodeId_empty_DB(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewUsersRepository(pool)

	found, err := repo.ByGithubNodeId("blah")

	assert.Nil(t, found)
	assert.Nil(t, err)
}

func Test_UsersRepository_ByGithubNodeId_finds_one(t *testing.T) {
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

func Test_UserRepository_Save_new_user(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewUsersRepository(pool)

	user, err := users.HydrateUser(
		"019a3f4d-978c-7e15-91ce-7dd2575485f5",
		"josephus miller",
		"josephus@starhelix.org.ceres",
		"some_node_id",
	)

	require.Nil(t, err)

	require.Nil(t, repo.Save(user))

	rows := getAllUsers(t, pool)

	require.Len(t, rows, 1, "expected exactly one row in users table")

	row := rows[0]
	assertUserEqualsRow(t, user, row)
}

func Test_UserRepository_Save_upsert_user(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewUsersRepository(pool)

	user, err := users.HydrateUser(
		"019a3f4d-978c-7e15-91ce-7dd2575485f5",
		"josephus miller",
		"josephus@starhelix.org.ceres",
		"some_node_id",
	)

	require.Nil(t, err)

	require.Nil(t, repo.Save(user))

	rows := getAllUsers(t, pool)

	require.Len(t, rows, 1, "expected exactly one row in users table")

	row := rows[0]

	assertUserEqualsRow(t, user, row)

	updatedUser, err := users.HydrateUser(
		"019a3f4d-978c-7e15-91ce-7dd2575485f5",
		"amos burton",
		"amos@thechurn.com",
		"other_node_id",
	)

	require.Nil(t, err)

	require.Nil(t, repo.Save(updatedUser))

	rows = getAllUsers(t, pool)

	require.Len(t, rows, 1, "expected exactly one row in users table")

	row = rows[0]

	assertUserEqualsRow(t, updatedUser, row)
}

func Test_UserRepository_Save_duped_email(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewUsersRepository(pool)

	user, err := users.HydrateUser(
		"019a3f4d-978c-7e15-91ce-7dd2575485f5",
		"josephus miller",
		"josephus@starhelix.org.ceres",
		"some_node_id",
	)

	require.Nil(t, err)

	require.Nil(t, repo.Save(user))

	rows := getAllUsers(t, pool)

	require.Len(t, rows, 1, "expected exactly one row in users table")

	row := rows[0]

	assertUserEqualsRow(t, user, row)

	newUser, err := users.HydrateUser(
		"019a3f82-7506-762f-bf96-221096985582",
		"joe miller",
		"josephus@starhelix.org.ceres",
		"some_node_id",
	)

	require.Nil(t, err)

	require.Equal(t, domain.ErrEmailAlreadyInUse{}, repo.Save(newUser))

	rows = getAllUsers(t, pool)

	require.Len(t, rows, 1, "expected exactly one row in users table")

	row = rows[0]

	assertUserEqualsRow(t, user, row)
}
