package domain

import (
	"github.com/google/uuid"
	"github.com/panoptescloud/api/internal/common/dto"
)

type Organisation struct {
	id      OrganisationID
	name    Name
	members Members
}

func (u *Organisation) ID() OrganisationID {
	return u.id
}

func (u *Organisation) Name() Name {
	return u.name
}

func (u *Organisation) Members() Members {
	return u.members
}

func NewOrganisation(id OrganisationID, name Name, members Members) *Organisation {
	return &Organisation{
		id:      id,
		name:    name,
		members: members,
	}
}

// HydrateOrganisation is designed to bypass any checks for invariants, typically
// used during tests or when hydrating from a DB.
func HydrateOrganisation(
	id string,
	name string,
	members Members,
) (*Organisation, error) {
	uid, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	return &Organisation{
		id: OrganisationID{
			value: uid,
		},
		name: Name{
			value: name,
		},
		members: members,
	}, nil
}

// --- APIKey
type APIKey struct {
	id             APIKeyID
	name           APIKeyName
	organisationID OrganisationID
	token          dto.HashedValue
}

func (key APIKey) ID() APIKeyID {
	return key.id
}

func (key APIKey) OrganisationID() OrganisationID {
	return key.organisationID
}

func (key APIKey) Name() APIKeyName {
	return key.name
}

func (key APIKey) Token() dto.HashedValue {
	return key.token
}

func NewAPIKey(id APIKeyID, orgID OrganisationID, name APIKeyName, token dto.HashedValue) *APIKey {
	return &APIKey{
		id:             id,
		name:           name,
		organisationID: orgID,
		token:          token,
	}
}

// HydrateAPIKey is designed to bypass any checks for invariants, typically
// used during tests or when hydrating from a DB.
func HydrateAPIKey(
	id string,
	name string,
	orgID string,
	encryptedTokenValue string,
) (*APIKey, error) {
	uid, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	orgUid, err := uuid.Parse(orgID)

	if err != nil {
		return nil, err
	}

	return &APIKey{
		id: APIKeyID{
			value: uid,
		},
		organisationID: OrganisationID{
			value: orgUid,
		},
		name: APIKeyName{
			value: name,
		},
		token: dto.HashedValue{
			Value: encryptedTokenValue,
		},
	}, nil
}
