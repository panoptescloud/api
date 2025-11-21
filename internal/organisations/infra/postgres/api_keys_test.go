package postgres_test

import (
	"testing"

	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/organisations/domain"
	"github.com/panoptescloud/api/internal/organisations/infra/postgres"
	dbassert "github.com/panoptescloud/api/tests/db/assert"
	"github.com/panoptescloud/api/tests/db/postgrestest"
	"github.com/panoptescloud/api/tests/db/seed"
	testutil "github.com/panoptescloud/api/tests/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_APIKeysRepository_GetByToken_NoRows(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewAPIKeysRepository(pool)

	key, err := repo.ByToken(dto.HashedValue{
		Value: "non-existent",
	})

	assert.Nil(t, err)
	assert.Nil(t, key)
}

func Test_APIKeysRepository_ByID(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	apiKeyID := testutil.NewUuidV7(t)
	userID := testutil.NewUuidV7(t)
	orgID := testutil.NewUuidV7(t)
	// These need to exist due to foreign keys
	_ = seed.User(pool, userID, "naomi-nagata", "naomia@opa.xyz", "some-node")
	_ = seed.Organisation(pool, orgID, "opa", userID, "owner")
	_ = seed.OrganisationAPIKey(pool, apiKeyID, orgID, "my-key", "some-secret")

	repo := postgres.NewAPIKeysRepository(pool)

	key, err := repo.ByToken(dto.HashedValue{
		Value: "some-secret",
	})

	require.Nil(t, err)
	assert.Equal(t, apiKeyID.String(), key.ID().String())
	assert.Equal(t, "my-key", key.Name().String())
	assert.Equal(t, "some-secret", key.Token().Value)
	assert.Equal(t, orgID.String(), key.OrganisationID().String())
}

func Test_APIKeysRepository_Save(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	apiKeyID := testutil.NewUuidV7(t)
	userID := testutil.NewUuidV7(t)
	orgID := testutil.NewUuidV7(t)
	// These need to exist due to foreign keys
	_ = seed.User(pool, userID, "naomi-nagata", "naomia@opa.xyz", "some-node")
	_ = seed.Organisation(pool, orgID, "opa", userID, "owner")

	repo := postgres.NewAPIKeysRepository(pool)

	key, err := domain.HydrateAPIKey(
		apiKeyID.String(),
		"some-key",
		orgID.String(),
		"some-secret-token",
	)

	require.Nil(t, err)

	err = repo.Save(key)

	assert.Nil(t, err)

	q := `SELECT 
		oak.id,
		oak.organisation_id,
		oak.token,
		oak.name
	FROM organisation_api_keys oak;
	`

	dbassert.RowsExist(t, pool, q, []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":              [16]uint8(apiKeyID),
					"name":            "some-key",
					"token":           "some-secret-token",
					"organisation_id": [16]uint8(orgID),
				},
			},
		},
	})
}

func Test_APIKeysRepository_Save_NameConflict(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	apiKeyID := testutil.NewUuidV7(t)
	userID := testutil.NewUuidV7(t)
	orgID := testutil.NewUuidV7(t)
	// These need to exist due to foreign keys
	_ = seed.User(pool, userID, "naomi-nagata", "naomia@opa.xyz", "some-node")
	_ = seed.Organisation(pool, orgID, "opa", userID, "owner")

	repo := postgres.NewAPIKeysRepository(pool)

	key, err := domain.HydrateAPIKey(
		apiKeyID.String(),
		"some-key",
		orgID.String(),
		"some-secret-token",
	)

	require.Nil(t, err)

	err = repo.Save(key)

	assert.Nil(t, err)

	q := `SELECT 
		oak.id,
		oak.organisation_id,
		oak.token,
		oak.name
	FROM organisation_api_keys oak;
	`

	dbassert.RowsExist(t, pool, q, []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":              [16]uint8(apiKeyID),
					"name":            "some-key",
					"token":           "some-secret-token",
					"organisation_id": [16]uint8(orgID),
				},
			},
		},
	})

	newID := testutil.NewUuidV7(t)
	newOrg, err := domain.HydrateAPIKey(newID.String(), "some-key", orgID.String(), "some-new-secret-token")

	require.Nil(t, err)
	err = repo.Save(newOrg)

	assert.Equal(t, domain.ErrAPIKeyNameAlreadyInUse{}, err)

	dbassert.RowsExist(t, pool, q, []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":              [16]uint8(apiKeyID),
					"name":            "some-key",
					"token":           "some-secret-token",
					"organisation_id": [16]uint8(orgID),
				},
			},
		},
	})
}

func Test_APIKeysRepository_Save_TokenConflict(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	apiKeyID := testutil.NewUuidV7(t)
	userID := testutil.NewUuidV7(t)
	orgID := testutil.NewUuidV7(t)
	// These need to exist due to foreign keys
	_ = seed.User(pool, userID, "naomi-nagata", "naomia@opa.xyz", "some-node")
	_ = seed.Organisation(pool, orgID, "opa", userID, "owner")

	repo := postgres.NewAPIKeysRepository(pool)

	key, err := domain.HydrateAPIKey(
		apiKeyID.String(),
		"some-key",
		orgID.String(),
		"some-secret-token",
	)

	require.Nil(t, err)

	err = repo.Save(key)

	assert.Nil(t, err)

	q := `SELECT 
		oak.id,
		oak.organisation_id,
		oak.token,
		oak.name
	FROM organisation_api_keys oak;
	`

	dbassert.RowsExist(t, pool, q, []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":              [16]uint8(apiKeyID),
					"name":            "some-key",
					"token":           "some-secret-token",
					"organisation_id": [16]uint8(orgID),
				},
			},
		},
	})

	newID := testutil.NewUuidV7(t)
	newOrg, err := domain.HydrateAPIKey(newID.String(), "some-different-key", orgID.String(), "some-secret-token")

	require.Nil(t, err)
	err = repo.Save(newOrg)

	assert.Equal(t, domain.ErrTokenConflict{}, err)

	dbassert.RowsExist(t, pool, q, []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":              [16]uint8(apiKeyID),
					"name":            "some-key",
					"token":           "some-secret-token",
					"organisation_id": [16]uint8(orgID),
				},
			},
		},
	})
}
