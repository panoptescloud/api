package domain

import (
	"github.com/google/uuid"
)

// --- MemberID
type MemberID struct {
	value uuid.UUID
}

func (id MemberID) String() string {
	return id.value.String()
}

func (id MemberID) Bytes() []byte {
	return id.value[:]
}

func (id MemberID) WrappedUuid() uuid.UUID {
	return id.value
}

type MemberRole struct {
	value string
}

func (mr MemberRole) String() string {
	return mr.value
}

// --- Member
type Member struct {
	id   MemberID
	role MemberRole
}

func (m Member) ID() MemberID {
	return m.id
}

func (m Member) Role() MemberRole {
	return m.role
}

func HydrateMember(id string, role string) (Member, error) {
	uid, err := uuid.Parse(id)

	if err != nil {
		return Member{}, err
	}

	return Member{
		id: MemberID{
			value: uid,
		},
		role: MemberRole{
			value: role,
		},
	}, nil
}

// --- Members
type Members []Member

// --- Name
type Name struct {
	value string
}

func (n Name) String() string {
	return n.value
}

// --- OrganisationID
type OrganisationID struct {
	value uuid.UUID
}

func (id OrganisationID) String() string {
	return id.value.String()
}

func (id OrganisationID) Bytes() []byte {
	return id.value[:]
}

func (id OrganisationID) WrappedUuid() uuid.UUID {
	return id.value
}

func GenerateOrganisationID() (OrganisationID, error) {
	id, err := uuid.NewV7()

	if err != nil {
		return OrganisationID{}, err
	}

	return OrganisationID{
		value: id,
	}, nil
}
