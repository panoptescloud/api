package users

import (
	"github.com/google/uuid"
)

type GithubIdentity struct {
	nodeID GithubUserNodeId
}

func (gi GithubIdentity) NodeID() GithubUserNodeId {
	return gi.nodeID
}

func NewGithubIdentity(nodeID GithubUserNodeId) GithubIdentity {
	return GithubIdentity{
		nodeID: nodeID,
	}
}

type User struct {
	id UserID
	name Name
	email Email
	githubIdentity GithubIdentity
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Name() Name {
	return u.name
}

func (u *User) Rename(new Name) {
	u.name = new
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) GithubIdentity() GithubIdentity {
	return u.githubIdentity
}

func NewUser(id UserID, name Name, email Email, githubIdentity GithubIdentity) *User {
	return &User{
		id: id, 
		name: name,
		email: email,
		githubIdentity: githubIdentity,
	}

	
}

// HydrateUser is designed to bypass any checks for invariants, typically used 
// during tests or when hydrating from a DB.
func HydrateUser(
	id string,
	name string,
	email string,
	githubIdentityNodeId string,
) (*User, error) {
	githubIdentity := GithubIdentity{
		nodeID: GithubUserNodeId{
			value: githubIdentityNodeId,
		},
	}

	uid, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	return &User{
		id: UserID{
			value: uid,
		},
		name: Name{
			value: name,
		},
		email: Email{
			value: email,
		},
		githubIdentity: githubIdentity,
	}, nil
}