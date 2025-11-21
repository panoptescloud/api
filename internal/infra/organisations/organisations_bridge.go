package organisations

import (
	"errors"

	"github.com/google/uuid"
	"github.com/panoptescloud/api/internal/common/bus"
	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/organisations/application"
	"github.com/panoptescloud/api/internal/organisations/domain"
)

type OrganisationsBridge struct {
	bus *bus.Bus
}

func (a *OrganisationsBridge) LoadByUserID(userID uuid.UUID) ([]dto.ActorOrganisation, error) {
	memberID := domain.NewMemberID(userID)
	orgs, err := bus.RunQuery[application.GetOrganisationsForMember, []*domain.Organisation](a.bus, application.GetOrganisationsForMember{
		ID: memberID,
	})

	if err != nil {
		return nil, err
	}

	ao := make([]dto.ActorOrganisation, len(orgs))

	for _, o := range orgs {
		member, found := o.Members().ByID(memberID)

		if !found {
			return nil, errors.New("member not found in organisation; this shouldn't logically be possible")
		}

		ao = append(ao, dto.NewActorOrganisation(
			o.ID().WrappedUuid(),
			member.Role().String(),
		))
	}

	return ao, nil
}

func NewOrganisationsBridge(b *bus.Bus) *OrganisationsBridge {
	return &OrganisationsBridge{
		bus: b,
	}
}
