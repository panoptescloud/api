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

func Test_APIKeysRepository_GetByToken(t *testing.T) {
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

func Test_APIKeysRepository_ByID_NoRows(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewAPIKeysRepository(pool)

	id := testutil.NewUuidV7(t)
	apiKeyID, err := domain.NewAPIKeyID(id.String())
	require.Nil(t, err)

	key, err := repo.ByID(apiKeyID)

	assert.Nil(t, err)
	assert.Nil(t, key)
}

func Test_APIKeysRepository_ByID(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	apiKeyID, err := domain.GenerateAPIKeyID()
	require.Nil(t, err)
	userID := testutil.NewUuidV7(t)
	orgID := testutil.NewUuidV7(t)
	// These need to exist due to foreign keys
	_ = seed.User(pool, userID, "naomi-nagata", "naomia@opa.xyz", "some-node")
	_ = seed.Organisation(pool, orgID, "opa", userID, "owner")
	_ = seed.OrganisationAPIKey(pool, apiKeyID.WrappedUuid(), orgID, "my-key", "some-secret")

	repo := postgres.NewAPIKeysRepository(pool)

	key, err := repo.ByID(apiKeyID)

	require.Nil(t, err)
	assert.Equal(t, apiKeyID.String(), key.ID().String())
	assert.Equal(t, "my-key", key.Name().String())
	assert.Equal(t, "some-secret", key.Token().Value)
	assert.Equal(t, orgID.String(), key.OrganisationID().String())
}

func Test_APIKeysRepository_AllForOrganisation_NoRows(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewAPIKeysRepository(pool)

	id, err := domain.GenerateOrganisationID()
	require.Nil(t, err)

	keys, err := repo.AllForOrganisation(id)

	assert.Nil(t, err)
	assert.Equal(t, []*domain.APIKey{}, keys)
}

func Test_APIKeysRepository_AllForOrganisation(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	apiKeyID, err := domain.GenerateAPIKeyID()
	require.Nil(t, err)
	apiKeyID2, err := domain.GenerateAPIKeyID()
	require.Nil(t, err)
	apiKeyID3, err := domain.GenerateAPIKeyID()
	require.Nil(t, err)
	apiKeyID4, err := domain.GenerateAPIKeyID()
	require.Nil(t, err)
	userID := testutil.NewUuidV7(t)
	userID2 := testutil.NewUuidV7(t)
	orgID, err := domain.GenerateOrganisationID()
	require.Nil(t, err)
	orgID2, err := domain.GenerateOrganisationID()
	require.Nil(t, err)
	// These need to exist due to foreign keys
	_ = seed.User(pool, userID, "naomi-nagata", "naomia@opa.xyz", "some-node")
	_ = seed.User(pool, userID2, "alex-kamal", "alexkamal@mcrn.org", "some-other-node")
	_ = seed.Organisation(pool, orgID.WrappedUuid(), "opa", userID, "owner")
	_ = seed.Organisation(pool, orgID.WrappedUuid(), "mcrn", userID2, "owner")
	_ = seed.OrganisationAPIKey(pool, apiKeyID.WrappedUuid(), orgID.WrappedUuid(), "my-key", "some-secret")
	_ = seed.OrganisationAPIKey(pool, apiKeyID2.WrappedUuid(), orgID.WrappedUuid(), "another-key", "another-secret")
	_ = seed.OrganisationAPIKey(pool, apiKeyID3.WrappedUuid(), orgID.WrappedUuid(), "and-another-key", "and-another-secret")
	_ = seed.OrganisationAPIKey(pool, apiKeyID4.WrappedUuid(), orgID2.WrappedUuid(), "org2-key", "org2-secret")

	repo := postgres.NewAPIKeysRepository(pool)

	keys, err := repo.AllForOrganisation(orgID)

	require.Nil(t, err)

	found1, err := domain.HydrateAPIKey(apiKeyID.String(), "my-key", orgID.String(), "some-secret")
	require.Nil(t, err)
	found2, err := domain.HydrateAPIKey(apiKeyID2.String(), "another-key", orgID.String(), "another-secret")
	require.Nil(t, err)
	found3, err := domain.HydrateAPIKey(apiKeyID3.String(), "and-another-key", orgID.String(), "and-another-secret")
	require.Nil(t, err)
	assert.Equal(t, []*domain.APIKey{
		found1,
		found2,
		found3,
	}, keys)
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
