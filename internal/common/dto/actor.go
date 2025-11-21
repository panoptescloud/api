package dto

import "github.com/google/uuid"

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
type Actor struct {
	userID uuid.UUID
}

func (a *Actor) UserID() uuid.UUID {
	return a.userID
}

// Think this is how we'll handle permissions, keep it on the actor, and we'll
// load permissions from the DB (once they exist) and check against the action.
// Think we'll make the value here something more complex so we can check things
// like can update organisation IF they're the owner and such
func (a *Actor) Can(action string) bool {
	return true
}

func NewActor(id uuid.UUID) *Actor {
	return &Actor{
		userID: id,
	}
}
