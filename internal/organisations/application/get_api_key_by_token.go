package application

import (
	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/organisations/domain"
)

type GetOrganisationAPIKeyByToken struct {
	Token string
}

func (cmd GetOrganisationAPIKeyByToken) GetName() string {
	return "organisations.command.get_organisation_api_key_by_id"
}

func (h *OrganisationQueryHandler) GetOrganisationAPIKeyByToken(cmd GetOrganisationAPIKeyByToken) (*domain.APIKey, error) {
	return h.apiKeyRepo.ByToken(dto.HashedValue{
		Value: cmd.Token,
	})
}
