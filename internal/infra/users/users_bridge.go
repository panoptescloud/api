package actors

import (
	"github.com/google/uuid"
	"github.com/panoptescloud/api/internal/common/bus"
	"github.com/panoptescloud/api/internal/common/dto"
	users "github.com/panoptescloud/api/internal/users/application"
	"github.com/panoptescloud/api/internal/users/domain"
)

type UsersBridge struct {
	bus *bus.Bus
}

func (a *UsersBridge) ByID(id uuid.UUID) (*dto.Actor, error) {
	userID, err := domain.NewUserID(id.String())

	if err != nil {
		return nil, err
	}

	u, err := bus.RunQuery[users.GetUserByID, *domain.User](a.bus, users.GetUserByID{
		ID: userID,
	})

	if err != nil {
		return nil, err
	}

	return dto.NewActor(u.ID().WrappedUuid()), nil
}

func NewUsersBridge(b *bus.Bus) *UsersBridge {
	return &UsersBridge{
		bus: b,
	}
}
