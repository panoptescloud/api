package postgres_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/panoptescloud/api/internal/common"
	usersdomain "github.com/panoptescloud/api/internal/users/domain"
	"github.com/panoptescloud/api/internal/users/infra/postgres"
	dbassert "github.com/panoptescloud/api/tests/db/assert"
	"github.com/panoptescloud/api/tests/db/postgrestest"
	"github.com/panoptescloud/api/tests/db/seed"
	testutil "github.com/panoptescloud/api/tests/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func allUsersQuery() string {
	return `
	SELECT
		u.id,
		u.name,
		u.email,
		gu.user_id as gh_user_id,
		gu.node_id as gh_node_id
	FROM users u
	INNER JOIN github_users gu 
	ON u.id=gu.user_id;
	`
}

func assertUserEqualsRow(t *testing.T, user *usersdomain.User, dbRow []any) {
	dbID := uuid.UUID{}
	idBytes := dbRow[0].([16]uint8)
	copy(dbID[:], idBytes[:])

	require.Equal(t, user.ID().String(), dbID.String())
	require.Equal(t, user.Name().String(), dbRow[1].(string))
	require.Equal(t, user.Email().String(), dbRow[2].(string))
	require.Equal(t, user.GithubIdentity().NodeID().String(), dbRow[3].(string))
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

	id := testutil.NewUuidV7(t)
	name := "josephus miller"
	email := "josephus@starhelix.org.ceres"
	nodeID := "some_node_id"
	err := seed.User(
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

	id := testutil.NewUuidV7(t)
	user, err := usersdomain.HydrateUser(
		id.String(),
		"josephus miller",
		"josephus@starhelix.org.ceres",
		"some_node_id",
	)

	require.Nil(t, err)

	require.Nil(t, repo.Save(user))

	dbassert.RowsExist(t, pool, allUsersQuery(), []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":         [16]uint8(id),
					"name":       "josephus miller",
					"email":      "josephus@starhelix.org.ceres",
					"gh_user_id": [16]uint8(id),
					"gh_node_id": "some_node_id",
				},
			},
		},
	})
}

func Test_UserRepository_Save_upsert_user(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewUsersRepository(pool)

	id := testutil.NewUuidV7(t)
	user, err := usersdomain.HydrateUser(
		id.String(),
		"josephus miller",
		"josephus@starhelix.org.ceres",
		"some_node_id",
	)

	require.Nil(t, err)

	require.Nil(t, repo.Save(user))

	dbassert.RowsExist(t, pool, allUsersQuery(), []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":         [16]uint8(id),
					"name":       "josephus miller",
					"email":      "josephus@starhelix.org.ceres",
					"gh_user_id": [16]uint8(id),
					"gh_node_id": "some_node_id",
				},
			},
		},
	})

	updatedUser, err := usersdomain.HydrateUser(
		id.String(),
		"amos burton",
		"amos@thechurn.com",
		"other_node_id",
	)

	require.Nil(t, err)

	require.Nil(t, repo.Save(updatedUser))

	dbassert.RowsExist(t, pool, allUsersQuery(), []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":         [16]uint8(id),
					"name":       "amos burton",
					"email":      "amos@thechurn.com",
					"gh_user_id": [16]uint8(id),
					"gh_node_id": "other_node_id",
				},
			},
		},
	})
}

func Test_UserRepository_Save_duped_email(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewUsersRepository(pool)

	id := testutil.NewUuidV7(t)
	user, err := usersdomain.HydrateUser(
		id.String(),
		"josephus miller",
		"josephus@starhelix.org.ceres",
		"some_node_id",
	)

	require.Nil(t, err)

	require.Nil(t, repo.Save(user))

	dbassert.RowsExist(t, pool, allUsersQuery(), []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":         [16]uint8(id),
					"name":       "josephus miller",
					"email":      "josephus@starhelix.org.ceres",
					"gh_user_id": [16]uint8(id),
					"gh_node_id": "some_node_id",
				},
			},
		},
	})

	newUser, err := usersdomain.HydrateUser(
		"019a3f82-7506-762f-bf96-221096985582",
		"joe miller",
		"josephus@starhelix.org.ceres",
		"some_node_id",
	)

	require.Nil(t, err)

	require.Equal(t, common.ErrEmailAlreadyInUse{}, repo.Save(newUser))

	dbassert.RowsExist(t, pool, allUsersQuery(), []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":         [16]uint8(id),
					"name":       "josephus miller",
					"email":      "josephus@starhelix.org.ceres",
					"gh_user_id": [16]uint8(id),
					"gh_node_id": "some_node_id",
				},
			},
		},
	})
}
