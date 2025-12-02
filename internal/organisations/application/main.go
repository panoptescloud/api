package application

import (
	"github.com/panoptescloud/api/internal/common/bus"
	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/organisations/domain"
)

type userBridge interface {
	UserExists(id domain.MemberID) (bool, error)
}

type organisationQueryRepo interface {
	ByID(id domain.OrganisationID) (*domain.Organisation, error)
	ForMember(id domain.MemberID) ([]*domain.Organisation, error)
}

type OrganisationQueryHandler struct {
	organisationRepo organisationQueryRepo
	apiKeyRepo       apiKeyRepo
}

func NewOrganisationQueryHandler(organisationRepo organisationQueryRepo, apiKeyRepo apiKeyRepo) *OrganisationQueryHandler {
	return &OrganisationQueryHandler{
		organisationRepo: organisationRepo,
		apiKeyRepo:       apiKeyRepo,
	}
}

type organisationValidationRepo interface {
	ByID(id domain.OrganisationID) (*domain.Organisation, error)
}

type organisationCmdRepo interface {
	organisationValidationRepo
	Save(*domain.Organisation) error
}

type apiKeyQueryRepo interface {
	ByID(domain.APIKeyID) (*domain.APIKey, error)
	ByToken(token dto.HashedValue) (*domain.APIKey, error)
	AllForOrganisation(domain.OrganisationID) ([]*domain.APIKey, error)
}

type apiKeyValidationRepo interface {
	// Nothing yet
}

type apiKeyCmdRepo interface {
	apiKeyValidationRepo
	Save(*domain.APIKey) error
}

type OrganisationCmdHandler struct {
	organisationRepo organisationCmdRepo
	apiKeyRepo       apiKeyCmdRepo
	userBridge       userBridge
}

func NewOrganisationCmdHandler(organisationRepo organisationCmdRepo, apiKeyRepo apiKeyCmdRepo, ub userBridge) *OrganisationCmdHandler {
	return &OrganisationCmdHandler{
		organisationRepo: organisationRepo,
		apiKeyRepo:       apiKeyRepo,
		userBridge:       ub,
	}
}

type organisationRepo interface {
	organisationQueryRepo
	organisationCmdRepo
}

type apiKeyRepo interface {
	apiKeyQueryRepo
	apiKeyCmdRepo
}

func RegisterToBus(b *bus.Bus, r organisationRepo, apiKeyRepo apiKeyRepo, ub userBridge) error {
	cmdHandler := NewOrganisationCmdHandler(r, apiKeyRepo, ub)
	queryHandler := NewOrganisationQueryHandler(r, apiKeyRepo)

	bus.RegisterCommand(b, cmdHandler.Create)
	bus.RegisterCommand(b, cmdHandler.CreateAPIKey)

	bus.RegisterQuery(b, queryHandler.GetByID)
	bus.RegisterQuery(b, queryHandler.GetAllForMember)
	bus.RegisterQuery(b, queryHandler.GetAPIKeyByID)
	bus.RegisterQuery(b, queryHandler.GetOrganisationAPIKeyByToken)
	bus.RegisterQuery(b, queryHandler.GetAPIKeysForOrganisation)

	return nil
}
