package application

import (
	"github.com/panoptescloud/api/internal/common/bus"
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
}

func NewOrganisationQueryHandler(organisationRepo organisationQueryRepo) *OrganisationQueryHandler {
	return &OrganisationQueryHandler{
		organisationRepo: organisationRepo,
	}
}

type organisationValidationRepo interface {
	// TODO: add methods required
}

type organisationCmdRepo interface {
	organisationValidationRepo
	Save(*domain.Organisation) error
}

type OrganisationCmdHandler struct {
	organisationRepo organisationCmdRepo
	userBridge       userBridge
}

func NewOrganisationCmdHandler(organisationRepo organisationCmdRepo, ub userBridge) *OrganisationCmdHandler {
	return &OrganisationCmdHandler{
		organisationRepo: organisationRepo,
		userBridge:       ub,
	}
}

type organisationRepo interface {
	organisationQueryRepo
	organisationCmdRepo
}

func RegisterToBus(b *bus.Bus, r organisationRepo, ub userBridge) error {
	cmdHandler := NewOrganisationCmdHandler(r, ub)
	queryHandler := NewOrganisationQueryHandler(r)

	bus.RegisterCommand(b, cmdHandler.Create)
	bus.RegisterQuery(b, queryHandler.GetByID)
	bus.RegisterQuery(b, queryHandler.GetAllForMember)

	return nil
}
