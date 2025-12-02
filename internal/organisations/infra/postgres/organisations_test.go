package postgres_test

import (
	"testing"

	"github.com/panoptescloud/api/internal/organisations/domain"
	"github.com/panoptescloud/api/internal/organisations/infra/postgres"
	dbassert "github.com/panoptescloud/api/tests/db/assert"
	"github.com/panoptescloud/api/tests/db/postgrestest"
	"github.com/panoptescloud/api/tests/db/seed"
	testutil "github.com/panoptescloud/api/tests/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_OrganisationsRepository_GetByID_NoRows(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	repo := postgres.NewOrganisationsRepository(pool)

	id, err := domain.GenerateOrganisationID()
	require.Nil(t, err)
	org, err := repo.ByID(id)

	assert.Nil(t, err)
	assert.Nil(t, org)
}

func Test_OrganisationsRepository_Save(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	userID := testutil.NewUuidV7(t)
	_ = seed.User(pool, userID, "naomi-nagata", "naomia@opa.xyz", "some-node")
	repo := postgres.NewOrganisationsRepository(pool)

	member, err := domain.HydrateMember(userID.String(), "admin")

	require.Nil(t, err)

	id := testutil.NewUuidV7(t)

	org, err := domain.HydrateOrganisation(id.String(), "my-org", domain.Members{
		member,
	})

	require.Nil(t, err)

	err = repo.Save(org)

	assert.Nil(t, err)

	q := `SELECT 
		o.id,
		o.name,
		om.role as member_role,
		om.organisation_id as member_org_id,
		om.member_id as member_id
	FROM organisations o
	INNER JOIN organisation_members om 
		ON o.id=om.organisation_id;
	`

	dbassert.RowsExist(t, pool, q, []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":            [16]uint8(id),
					"name":          "my-org",
					"member_org_id": [16]uint8(id),
					"member_id":     [16]uint8(userID),
					"member_role":   "admin",
				},
			},
		},
	})
}

func Test_OrganisationsRepository_Save_NameConflict(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	userID := testutil.NewUuidV7(t)
	_ = seed.User(pool, userID, "naomi-nagata", "naomia@opa.xyz", "some-node")
	repo := postgres.NewOrganisationsRepository(pool)

	member, err := domain.HydrateMember(userID.String(), "admin")

	require.Nil(t, err)

	id := testutil.NewUuidV7(t)

	org, err := domain.HydrateOrganisation(id.String(), "my-org", domain.Members{
		member,
	})

	require.Nil(t, err)

	err = repo.Save(org)

	assert.Nil(t, err)

	q := `SELECT 
		o.id,
		o.name,
		om.role as member_role,
		om.organisation_id as member_org_id,
		om.member_id as member_id
	FROM organisations o
	INNER JOIN organisation_members om 
		ON o.id=om.organisation_id;
	`

	dbassert.RowsExist(t, pool, q, []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":            [16]uint8(id),
					"name":          "my-org",
					"member_org_id": [16]uint8(id),
					"member_id":     [16]uint8(userID),
					"member_role":   "admin",
				},
			},
		},
	})

	newID := testutil.NewUuidV7(t)
	newOrg, err := domain.HydrateOrganisation(newID.String(), "my-org", domain.Members{
		member,
	})

	require.Nil(t, err)
	err = repo.Save(newOrg)

	assert.Equal(t, domain.ErrNameAlreadyInUse{}, err)

	dbassert.RowsExist(t, pool, q, []dbassert.Expectation{
		dbassert.RowsExpectation{
			Rows: []map[string]any{
				{
					"id":            [16]uint8(id),
					"name":          "my-org",
					"member_org_id": [16]uint8(id),
					"member_id":     [16]uint8(userID),
					"member_role":   "admin",
				},
			},
		},
	})
}

func Test_OrganisationsRepository_ByID(t *testing.T) {
	pool := postgrestest.PrepareDBForTest(t, pool, true)
	defer pool.Close()

	orgID, err := domain.GenerateOrganisationID()
	require.Nil(t, err)

	userID := testutil.NewUuidV7(t)
	err = seed.User(pool, userID, "naomi-nagata", "naomia@opa.xyz", "some-node")
	require.Nil(t, err, "failed to seed user")

	err = seed.Organisation(pool, orgID.WrappedUuid(), "my-org", userID, "admin")
	require.Nil(t, err, "failed to seed organisation")

	repo := postgres.NewOrganisationsRepository(pool)
	org, err := repo.ByID(orgID)

	assert.Nil(t, err)
	assert.Equal(t, orgID, org.ID())
	assert.Equal(t, "my-org", org.Name().String())
	assert.Equal(t, userID.String(), org.Members()[0].ID().String())
	assert.Equal(t, "admin", org.Members()[0].Role().String())

}
