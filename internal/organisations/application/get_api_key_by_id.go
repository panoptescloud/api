package application

import (
	"github.com/panoptescloud/api/internal/organisations/domain"
)

type GetOrganisationAPIKeyByID struct {
	ID domain.APIKeyID
}

func (cmd GetOrganisationAPIKeyByID) GetName() string {
	return "organisations.command.get_organisation_api_key_by_id"
}

func (h *OrganisationQueryHandler) GetAPIKeyByID(dto GetOrganisationAPIKeyByID) (*domain.APIKey, error) {
	return h.apiKeyRepo.ByID(dto.ID)
}
