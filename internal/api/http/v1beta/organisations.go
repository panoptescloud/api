package v1beta

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/common"
	"github.com/panoptescloud/api/internal/common/bus"
	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/common/validation"
	"github.com/panoptescloud/api/internal/organisations/application"
	"github.com/panoptescloud/api/internal/organisations/domain"
)

type OrganisationMember struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type Organisation struct {
	ID      string               `json:"id"`
	Name    string               `json:"name"`
	Members []OrganisationMember `json:"members"`
}

type OrganisationsController struct {
	bus *bus.Bus
}

func (c *OrganisationsController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID:   "v1beta.organisations.create",
		Method:        http.MethodPost,
		Path:          "/organisations",
		Summary:       "Create an organisation",
		DefaultStatus: http.StatusOK,
		Tags: []string{
			"Organisations",
		},
		Metadata: map[string]any{},
	}, ErrorHandler(debugErrorsEnabled, c.Create))

	huma.Register(api, huma.Operation{
		OperationID:   "v1beta.organisations.get",
		Method:        http.MethodGet,
		Path:          "/organisations/:id",
		Summary:       "Get an organisation by ID.",
		DefaultStatus: http.StatusOK,
		Tags: []string{
			"Organisations",
		},
		Metadata: map[string]any{},
	}, ErrorHandler(debugErrorsEnabled, c.Get))

	huma.Register(api, huma.Operation{
		OperationID:   "v1beta.organisations.list",
		Method:        http.MethodGet,
		Path:          "/organisations",
		Summary:       "List all organisations.",
		DefaultStatus: http.StatusOK,
		Tags: []string{
			"Organisations",
		},
		Metadata: map[string]any{},
	}, ErrorHandler(debugErrorsEnabled, c.List))
}

func NewOrganisationsController(b *bus.Bus) *OrganisationsController {
	return &OrganisationsController{
		bus: b,
	}
}

type CreateOrganisationRequestBody struct {
	Name string `json:"name"`
}

type CreateOrganisationRequest struct {
	Body CreateOrganisationRequestBody
}

type CreateOrganisationResponse struct {
	Status int
	Body   Organisation
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

	memberID, err := domain.NewMemberID(actor.UserID().String())

	if err != nil {
		return nil, err
	}

	err = bus.Dispatch(c.bus, application.CreateOrganisation{
		ID:    id,
		Name:  req.Body.Name,
		Owner: memberID,
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
		Body: Organisation{
			ID:      org.ID().String(),
			Name:    org.Name().String(),
			Members: members,
		},
	}, nil
}

type GetOrganisationRequest struct {
	ID string `path:"id" doc:"UUID of the organisation."`
}

type GetOrganisationResponse struct {
	Status int
	Body   Organisation
}

func (c *OrganisationsController) Get(ctx context.Context, req *GetOrganisationRequest) (*GetOrganisationResponse, error) {
	id, err := domain.NewOrganisationID(req.ID)

	if err != nil {
		return nil, validation.Error{
			FieldErrors: []validation.FieldError{
				{
					Key: "id",
					Errors: validation.Violations{
						validation.InvalidUUIDViolation{},
					},
				},
			},
		}
	}

	org, err := bus.RunQuery[application.GetOrganisationByID, *domain.Organisation](c.bus, application.GetOrganisationByID{
		ID: id,
	})

	if err != nil {
		return nil, err
	}

	if org == nil {
		return nil, common.ErrResourceNotFound{}
	}

	members := make([]OrganisationMember, len(org.Members()))

	for i, m := range org.Members() {
		members[i] = OrganisationMember{
			ID:   m.ID().String(),
			Role: m.Role().String(),
		}
	}

	return &GetOrganisationResponse{
		Status: http.StatusOK,
		Body: Organisation{
			ID:      org.ID().String(),
			Name:    org.Name().String(),
			Members: members,
		},
	}, nil
}

type ListOrganisationsRequest struct{}

type ListOrganisationsResponse struct {
	Status int
	Body   []Organisation
}

func (c *OrganisationsController) List(ctx context.Context, req *ListOrganisationsRequest) (*ListOrganisationsResponse, error) {
	actorValue := ctx.Value("actor")

	actor, ok := actorValue.(*dto.Actor)

	if !ok {
		return nil, errors.New("failed to load actor, could not assert type")
	}

	if actor.Type() != dto.UserActor {
		return nil, common.ErrUnauthorised{}
	}

	id, err := domain.NewMemberID(actor.UserID().String())

	if err != nil {
		return nil, err
	}

	orgs, err := bus.RunQuery[application.GetOrganisationsForMember, []*domain.Organisation](c.bus, application.GetOrganisationsForMember{
		ID: id,
	})

	if err != nil {
		return nil, err
	}

	apiOrgs := make([]Organisation, len(orgs))

	for i, o := range orgs {
		members := make([]OrganisationMember, len(o.Members()))

		for i, m := range o.Members() {
			members[i] = OrganisationMember{
				ID:   m.ID().String(),
				Role: m.Role().String(),
			}
		}

		apiOrgs[i] = Organisation{
			ID:      o.ID().String(),
			Name:    o.Name().String(),
			Members: members,
		}
	}

	return &ListOrganisationsResponse{
		Status: http.StatusOK,
		Body:   apiOrgs,
	}, nil
}
