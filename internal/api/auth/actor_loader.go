package auth

import (
	"github.com/google/uuid"
	"github.com/panoptescloud/api/internal/common/dto"
)

type organisationsBridge interface {
	LoadByUserID(userID uuid.UUID) ([]dto.ActorOrganisation, error)
}

type usersBridge interface {
	ByID(id uuid.UUID) (*dto.Actor, error)
}

type ActorLoader struct {
	orgsBridge  organisationsBridge
	usersBridge usersBridge
}

func (al *ActorLoader) ByUserID(id uuid.UUID) (*dto.Actor, error) {
	return al.usersBridge.ByID(id)
}

func (al *ActorLoader) ByApiKeyID(id uuid.UUID) (*dto.Actor, error) {
	id, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	return dto.NewAPIKeyActor(id), nil
}

func NewActorLoader(orgsBridge organisationsBridge, usersBridge usersBridge) *ActorLoader {
	return &ActorLoader{
		orgsBridge:  orgsBridge,
		usersBridge: usersBridge,
	}
}
