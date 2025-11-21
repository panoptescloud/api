package application

import (
	"github.com/panoptescloud/api/internal/organisations/domain"
)

type GetAPIKeysForOrganisation struct {
	ID domain.OrganisationID
}

func (cmd GetAPIKeysForOrganisation) GetName() string {
	return "organisations.command.get_api_keys_for_organisation"
}

func (h *OrganisationQueryHandler) GetAPIKeysForOrganisation(dto GetAPIKeysForOrganisation) ([]*domain.APIKey, error) {
	return h.apiKeyRepo.AllForOrganisation(dto.ID)
}
