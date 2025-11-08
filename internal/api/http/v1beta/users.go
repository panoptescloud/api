package v1beta

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/api/http/operations"
)

type MeRequest struct{}

type MeResponseBody struct {
	ID string `json:"id"`
}

type MeResponse struct {
	Status int
	Body   MeResponseBody
}

type UsersController struct{}

func (c *UsersController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID:   "v1.users.me",
		Method:        http.MethodGet,
		Path:          "/users/me",
		Summary:       "Get information about yourself.",
		DefaultStatus: http.StatusOK,
		Tags: []string{
			"Users",
		},
		Metadata: map[string]any{
			operations.OptDisableAllDefaultResponses: true,
		},
	}, ErrorHandler(debugErrorsEnabled, c.Me))
}

func NewUsersController() *UsersController {
	return &UsersController{}
}

func (c *UsersController) Me(ctx context.Context, req *StartupRequest) (*MeResponse, error) {
	return &MeResponse{
		Status: http.StatusOK,
		Body: MeResponseBody{
			ID: "blah",
		},
	}, nil
}
