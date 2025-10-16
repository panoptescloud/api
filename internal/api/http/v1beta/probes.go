package v1beta

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/api/http/operations"
	"github.com/panoptescloud/api/internal/api/http/v1beta/responses"
)

type StartupRequest struct{}

type ProbesController struct{}

func (c *ProbesController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID:   "v1.probes.startup",
		Method:        http.MethodGet,
		Path:          "/_probes/startup",
		Summary:       "Check if the app is started up",
		DefaultStatus: http.StatusNoContent,
		Metadata: map[string]any{
			operations.OptDisableAllDefaults: true,
		},
	}, ErrorHandler(debugErrorsEnabled, c.Startup))
}

func NewProbesController() *ProbesController {
	return &ProbesController{}
}

func (c *ProbesController) Startup(ctx context.Context, req *StartupRequest) (*responses.NoContent, error) {
	return &responses.NoContent{
		Status: http.StatusNoContent,
	}, nil
}
