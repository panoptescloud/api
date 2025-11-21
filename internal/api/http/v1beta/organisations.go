package v1beta

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/common/bus"
	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/organisations/application"
	"github.com/panoptescloud/api/internal/organisations/domain"
)

type OrganisationMember struct {
	ID   string
	Role string
}

type CreateOrganisationRequestBody struct {
	Name string `json:"name"`
}

type CreateOrganisationRequest struct {
	Body CreateOrganisationRequestBody
}

type CreateOrganisationResponseBody struct {
	ID      string               `json:"id"`
	Name    string               `json:"name"`
	Members []OrganisationMember `json:"members"`
}

type CreateOrganisationResponse struct {
	Status int
	Body   CreateOrganisationResponseBody
}

type OrganisationsController struct {
	bus *bus.Bus
}

func (c *OrganisationsController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID:   "v1.organisations.create",
		Method:        http.MethodPost,
		Path:          "/organisations",
		Summary:       "Create an organisation",
		DefaultStatus: http.StatusOK,
		Tags: []string{
			"Organisations",
		},
		Metadata: map[string]any{},
	}, ErrorHandler(debugErrorsEnabled, c.Create))
}

func NewOrganisationsController(b *bus.Bus) *OrganisationsController {
	return &OrganisationsController{
		bus: b,
	}
}

func (c *OrganisationsController) Create(ctx context.Context, req *CreateOrganisationRequest) (*CreateOrganisationResponse, error) {
	actorValue := ctx.Value("actor")

	actor, ok := actorValue.(*dto.Actor)

	if !ok {
		return nil, errors.New("failed to load actor, could not assert type")
	}

	id, err := domain.GenerateOrganisationID()
	if err != nil {
		return nil, err
	}

	err = bus.Dispatch(c.bus, application.CreateOrganisation{
		ID:    id,
		Name:  req.Body.Name,
		Owner: domain.NewMemberID(*actor.UserID()),
	})

	if err != nil {
		return nil, err
	}

	org, err := bus.RunQuery[application.GetOrganisationByID, *domain.Organisation](c.bus, application.GetOrganisationByID{
		ID: id,
	})

	if err != nil {
		return nil, err
	}

	members := make([]OrganisationMember, len(org.Members()))

	for i, m := range org.Members() {
		members[i] = OrganisationMember{
			ID:   m.ID().String(),
			Role: m.Role().String(),
		}
	}

	return &CreateOrganisationResponse{
		Status: http.StatusOK,
		Body: CreateOrganisationResponseBody{
			ID:      org.ID().String(),
			Name:    org.Name().String(),
			Members: members,
		},
	}, nil
}
