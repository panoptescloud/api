package auth

import (
	"github.com/google/uuid"
	"github.com/panoptescloud/api/internal/common/dto"
)

type organisationsBridge interface {
	LoadByUserID(userID uuid.UUID) ([]dto.ActorOrganisation, error)
	GetApiKeyByToken(token string) (*dto.Actor, error)
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

func (al *ActorLoader) ByApiKey(token string) (*dto.Actor, error) {
	apiKey, err := al.orgsBridge.GetApiKeyByToken(token)

	if err != nil {
		return nil, err
	}

	if apiKey != nil {
		return apiKey, nil
	}

	// TODO: later will authorise against user api keys as well as org api keys
	return nil, nil
}

func NewActorLoader(orgsBridge organisationsBridge, usersBridge usersBridge) *ActorLoader {
	return &ActorLoader{
		orgsBridge:  orgsBridge,
		usersBridge: usersBridge,
	}
}
