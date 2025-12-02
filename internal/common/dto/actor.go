package dto

import "github.com/google/uuid"

type ActorType string

var (
	OrganisationAPIKeyActor ActorType = "organisation_api_key"
	UserActor               ActorType = "user"
)

// --- Actor
type ActorOrganisation struct {
	id   uuid.UUID
	role string
}

func NewActorOrganisation(id uuid.UUID, role string) ActorOrganisation {
	return ActorOrganisation{
		id:   id,
		role: role,
	}
}

// --- Actor
// Must always have a userID and ActorType
// apiKeyID may be set if the ActorType is ApiKeyActor
type Actor struct {
	userID    *uuid.UUID
	apiKeyID  *uuid.UUID
	actorType ActorType
}

func (a *Actor) UserID() *uuid.UUID {
	return a.userID
}

func (a *Actor) Type() ActorType {
	return a.actorType
}

func (a *Actor) ApiKeyID() *uuid.UUID {
	return a.apiKeyID
}

// Think this is how we'll handle permissions, keep it on the actor, and we'll
// load permissions from the DB (once they exist) and check against the action.
// Think we'll make the value here something more complex so we can check things
// like can update organisation IF they're the owner and such
func (a *Actor) Can(action string) bool {
	return true
}

func NewUserActor(id uuid.UUID) *Actor {
	return &Actor{
		userID:    &id,
		actorType: UserActor,
	}
}

func NewOrganisationAPIKeyActor(id uuid.UUID) *Actor {
	return &Actor{
		apiKeyID:  &id,
		actorType: OrganisationAPIKeyActor,
	}
}
