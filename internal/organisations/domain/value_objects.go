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

func NewMemberID(id string) (MemberID, error) {
	uid, err := uuid.Parse(id)

	if err != nil {
		return MemberID{}, err
	}

	return MemberID{
		value: uid,
	}, nil
}

// --- MemberRole
type MemberRole struct {
	value string
}

func (mr MemberRole) String() string {
	return mr.value
}

func NewMemberRole(role string) MemberRole {
	return MemberRole{
		value: role,
	}
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

func NewMember(id MemberID, role MemberRole) Member {
	return Member{
		id:   id,
		role: role,
	}
}

// --- Members
type Members []Member

func (m Members) ByID(id MemberID) (Member, bool) {
	for _, member := range m {
		if member.ID() == id {
			return member, true
		}
	}

	return Member{}, false
}

// --- Name
type Name struct {
	value string
}

func (n Name) String() string {
	return n.value
}

func NewName(name string) Name {
	return Name{
		value: name,
	}
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

func NewOrganisationID(v string) (OrganisationID, error) {
	id, err := uuid.Parse(v)

	if err != nil {
		return OrganisationID{}, err
	}

	return OrganisationID{
		value: id,
	}, nil

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

// --- APIKeyID
type APIKeyID struct {
	value uuid.UUID
}

func (id APIKeyID) String() string {
	return id.value.String()
}

func (id APIKeyID) Bytes() []byte {
	return id.value[:]
}

func (id APIKeyID) WrappedUuid() uuid.UUID {
	return id.value
}

func NewAPIKeyID(v string) (APIKeyID, error) {
	id, err := uuid.Parse(v)

	if err != nil {
		return APIKeyID{}, err
	}

	return APIKeyID{
		value: id,
	}, nil

}

func GenerateAPIKeyID() (APIKeyID, error) {
	id, err := uuid.NewV7()

	if err != nil {
		return APIKeyID{}, err
	}

	return APIKeyID{
		value: id,
	}, nil
}

// --- APIKeyName
type APIKeyName struct {
	value string
}

func (n APIKeyName) String() string {
	return n.value
}

func NewAPIKeyName(name string) APIKeyName {
	return APIKeyName{
		value: name,
	}
}
