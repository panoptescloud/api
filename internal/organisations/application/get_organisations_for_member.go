package application

import (
	"github.com/panoptescloud/api/internal/organisations/domain"
)

type GetOrganisationsForMember struct {
	ID domain.MemberID
}

func (cmd GetOrganisationsForMember) GetName() string {
	return "users.command.get_organisations_for_member"
}

func (h *OrganisationQueryHandler) GetAllForMember(dto GetOrganisationsForMember) ([]*domain.Organisation, error) {
	return h.organisationRepo.ForMember(dto.ID)
}
