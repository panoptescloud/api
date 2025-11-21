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

func (al *ActorLoader) ById(id uuid.UUID) (*dto.Actor, error) {
	return al.usersBridge.ByID(id)
}

func NewActorLoader(orgsBridge organisationsBridge, usersBridge usersBridge) *ActorLoader {
	return &ActorLoader{
		orgsBridge:  orgsBridge,
		usersBridge: usersBridge,
	}
}
