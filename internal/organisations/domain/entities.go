package domain

import "github.com/google/uuid"

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
