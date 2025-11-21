package v1beta

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/common"
	"github.com/panoptescloud/api/internal/common/bus"
	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/common/validation"
	"github.com/panoptescloud/api/internal/organisations/application"
	"github.com/panoptescloud/api/internal/organisations/domain"
	"github.com/panoptescloud/api/pkg/util/random"
)

type apiKeyHasher interface {
	Hash(input string) (dto.HashedValue, error)
}

type NewOrganisationAPIKey struct {
	OrganisationAPIKey
	Token string `json:"token"`
}

type OrganisationAPIKey struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganisationID string `json:"organisation_id"`
}

type OrganisationAPIKeysController struct {
	bus    *bus.Bus
	hasher apiKeyHasher
}

func (c *OrganisationAPIKeysController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID:   "v1beta.organisations.api_keys.create",
		Method:        http.MethodPost,
		Path:          "/organisations/:organisation_id/api_keys",
		Summary:       "Create an organisation API key",
		DefaultStatus: http.StatusOK,
		Tags: []string{
			"Organisations",
			"API Keys",
		},
		Metadata: map[string]any{},
	}, ErrorHandler(debugErrorsEnabled, c.Create))

	huma.Register(api, huma.Operation{
		OperationID:   "v1beta.organisations.api_keys.get",
		Method:        http.MethodGet,
		Path:          "/organisations/:organisation_id/api_keys/:id",
		Summary:       "Get an organisation API Key by ID.",
		DefaultStatus: http.StatusOK,
		Tags: []string{
			"Organisations",
			"API Keys",
		},
		Metadata: map[string]any{},
	}, ErrorHandler(debugErrorsEnabled, c.Get))

	huma.Register(api, huma.Operation{
		OperationID:   "v1beta.organisations.api_keys.list",
		Method:        http.MethodGet,
		Path:          "/organisations/:organisation_id/api_keys",
		Summary:       "List all organisation API Keys.",
		DefaultStatus: http.StatusOK,
		Tags: []string{
			"Organisations",
			"API Keys",
		},
		Metadata: map[string]any{},
	}, ErrorHandler(debugErrorsEnabled, c.List))
}

func NewOrganisationAPIKeysController(b *bus.Bus, apiKeyHasher apiKeyHasher) *OrganisationAPIKeysController {
	return &OrganisationAPIKeysController{
		bus:    b,
		hasher: apiKeyHasher,
	}
}

type CreateOrganisationAPIKeyRequestBody struct {
	Name string `json:"name"`
}

type CreateOrganisationAPIKeyRequest struct {
	OrganisationID string `path:"organisation_id" doc:"UUID of the organisation."`
	Body           CreateOrganisationRequestBody
}

type CreateOrganisationAPIKeyResponse struct {
	Status int
	Body   NewOrganisationAPIKey
}

func (c *OrganisationAPIKeysController) Create(ctx context.Context, req *CreateOrganisationAPIKeyRequest) (*CreateOrganisationAPIKeyResponse, error) {
	apiKeyID, err := domain.GenerateAPIKeyID()

	if err != nil {
		return nil, err
	}

	orgID, err := domain.NewOrganisationID(req.OrganisationID)

	if err != nil {
		return nil, err
	}

	randomToken, err := random.GenerateString()
	if err != nil {
		return nil, err
	}

	token, err := c.hasher.Hash(randomToken)

	if err != nil {
		return nil, err
	}

	err = bus.Dispatch(c.bus, application.CreateAPIKey{
		ID:             apiKeyID,
		OrganisationID: orgID,
		Name:           req.Body.Name,
		Token:          token,
	})

	if err != nil {
		return nil, err
	}

	key, err := bus.RunQuery[application.GetOrganisationAPIKeyByID, *domain.APIKey](c.bus, application.GetOrganisationAPIKeyByID{
		ID: apiKeyID,
	})

	if err != nil {
		return nil, err
	}

	return &CreateOrganisationAPIKeyResponse{
		Status: http.StatusCreated,
		Body: NewOrganisationAPIKey{
			OrganisationAPIKey: OrganisationAPIKey{
				ID:             key.ID().String(),
				OrganisationID: key.OrganisationID().String(),
				Name:           key.Name().String(),
			},
			Token: token.Value,
		},
	}, nil
}

type GetOrganisationAPIKeyRequest struct {
	OrganisationID string `path:"organisation_id" doc:"UUID of the organisation."`
	ID             string `path:"id" doc:"UUID of the API key."`
}

type GetOrganisationAPIKeyResponse struct {
	Status int
	Body   OrganisationAPIKey
}

func (c *OrganisationAPIKeysController) Get(ctx context.Context, req *GetOrganisationAPIKeyRequest) (*GetOrganisationAPIKeyResponse, error) {
	apiKeyID, err := domain.NewAPIKeyID(req.ID)

	if err != nil {
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
	}

	key, err := bus.RunQuery[application.GetOrganisationAPIKeyByID, *domain.APIKey](c.bus, application.GetOrganisationAPIKeyByID{
		ID: apiKeyID,
	})

	if err != nil {
		return nil, err
	}

	if key.OrganisationID().String() != req.OrganisationID {
		return nil, common.ErrResourceNotFound{}
	}

	return &GetOrganisationAPIKeyResponse{
		Status: http.StatusCreated,
		Body: OrganisationAPIKey{
			ID:             key.ID().String(),
			OrganisationID: key.OrganisationID().String(),
			Name:           key.Name().String(),
		},
	}, nil

}

type ListOrganisationAPIKeysRequest struct {
	OrganisationID string `path:"organisation_id" doc:"UUID of the organisation."`
}

type ListOrganisationAPIKeysResponse struct {
	Status int
	Body   []OrganisationAPIKey
}

func (c *OrganisationAPIKeysController) List(ctx context.Context, req *ListOrganisationAPIKeysRequest) (*ListOrganisationAPIKeysResponse, error) {
	orgID, err := domain.NewOrganisationID(req.OrganisationID)

	if err != nil {
		if err != nil {
			return nil, validation.Error{
				FieldErrors: []validation.FieldError{
					{
						Key: "organisation_id",
						Errors: validation.Violations{
							validation.InvalidUUIDViolation{},
						},
					},
				},
			}
		}
	}

	keys, err := bus.RunQuery[application.GetAPIKeysForOrganisation, []*domain.APIKey](c.bus, application.GetAPIKeysForOrganisation{
		ID: orgID,
	})

	if err != nil {
		return nil, err
	}

	responseKeys := make([]OrganisationAPIKey, len(keys))

	for i, k := range keys {
		responseKeys[i] = OrganisationAPIKey{
			ID:             k.ID().String(),
			OrganisationID: k.OrganisationID().String(),
			Name:           k.Name().String(),
		}
	}

	return &ListOrganisationAPIKeysResponse{
		Status: http.StatusOK,
		Body:   responseKeys,
	}, nil
}
