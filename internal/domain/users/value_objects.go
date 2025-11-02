package users

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// --- UserID
type UserID struct {
	value uuid.UUID
}

func (id UserID) String() string {
	return id.value.String()
}

func (id UserID) Bytes() []byte {
	return id.value[:]
}

func (id UserID) WrappedUuid() uuid.UUID {
	return id.value
}

func GenerateUserID() (UserID, error) {
	id, err := uuid.NewV7()

	if err != nil {
		return UserID{}, err
	}

	return UserID{
		value: id,
	}, nil
}

func NewUserID(v string) (UserID, error) {
	id, err := uuid.Parse(v)

	if err != nil {
		return UserID{}, err
	}

	return UserID{
		value: id,
	}, nil
}


// --- Email
type Email struct {
	value string
}

func (e Email) String() string {
	return e.value
}

// Basic validation for now
func NewEmail(value string) (Email, error) {
	if len(value) < 5 || !strings.Contains(value, "@") {
		return Email{}, errors.New("invalid email address")
	}

	return Email{value: value}, nil
}


// --- Name
type Name struct {
	value string
}

func (e Name) String() string {
	return e.value
}

// No validation for now
func NewName(value string) (Name) {
	return Name{
		value: value,
	}
}


// --- GithubUserNodeId
type GithubUserNodeId struct {
	value string
}

func (e GithubUserNodeId) String() string {
	return e.value
}

// No validation for now
func NewGithubUserNodeId(value string) (GithubUserNodeId) {
	return GithubUserNodeId{
		value: value,
	}
}