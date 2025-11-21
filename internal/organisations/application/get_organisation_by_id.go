package application

import (
	"github.com/panoptescloud/api/internal/organisations/domain"
)

type GetOrganisationByID struct {
	ID domain.OrganisationID
}

func (cmd GetOrganisationByID) GetName() string {
	return "users.command.get_organisation_by_id"
}

func (h *OrganisationQueryHandler) GetByID(dto GetOrganisationByID) (*domain.Organisation, error) {
	return h.organisationRepo.ByID(dto.ID)
}
