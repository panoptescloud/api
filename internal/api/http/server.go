package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/panoptescloud/api/internal/api/http/operations"
)

type Controller interface {
	RegisterRoutes(g huma.API, debugErrorsEnabled bool)
}

var opsWithoutBodies = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodOptions,
	http.MethodDelete,

	// Probably don't need this one, but leaving for good measure
	http.MethodTrace,
}

func addValidationErrorResponse(op *huma.Operation) {
	validationStatus := strconv.Itoa(http.StatusUnprocessableEntity)

	if slices.Contains(opsWithoutBodies, op.Method) {
		validationStatus = strconv.Itoa(http.StatusBadRequest)
	}

	if _, ok := op.Responses[validationStatus]; ok {
		return
	}

	op.Responses[validationStatus] = &huma.Response{
		Description: "validation error",
		Content: map[string]*huma.MediaType{
			"application/problem+json": {
				Schema: &huma.Schema{
					Ref: "#/components/schemas/ErrorModel",
				},
			},
		},
	}
}
func addInternalErrorResponse(op *huma.Operation) {
	internalError := strconv.Itoa(http.StatusInternalServerError)

	if _, ok := op.Responses[internalError]; ok {
		return
	}

	op.Responses[internalError] = &huma.Response{
		Description: "Internal server error",
		Content: map[string]*huma.MediaType{
			"application/problem+json": {
				Schema: &huma.Schema{
					Ref: "#/components/schemas/ErrorModel",
				},
			},
		},
	}
}

func addNotFoundResponse(op *huma.Operation) {
	var notFoundEnabled = true

	if v, ok := op.Metadata[operations.OptDisableNotFound]; ok {
		if optAsBool, ok := v.(bool); ok {
			notFoundEnabled = !optAsBool
		}
	}

	if !notFoundEnabled {
		return
	}

	notFound := strconv.Itoa(http.StatusNotFound)

	if _, ok := op.Responses[notFound]; ok {
		return
	}

	op.Responses[notFound] = &huma.Response{
		Description: "Resource Not Found",
		Content: map[string]*huma.MediaType{
			"application/problem+json": {
				Schema: &huma.Schema{
					Ref: "#/components/schemas/ErrorModel",
				},
			},
		},
	}
}

func configureDefaultResponses(api *huma.OpenAPI, op *huma.Operation) {

	if _, ok := op.Responses["default"]; ok {
		// Remove the default as it's an error, but has no status code
		// Maybe there's another way to turn it off
		op.Responses["default"] = nil
	}

	if v, ok := op.Metadata[operations.OptDisableAllDefaults]; ok {
		if optAsBool, ok := v.(bool); ok && optAsBool {
			return
		}
	}

	addValidationErrorResponse(op)
	addInternalErrorResponse(op)
	addNotFoundResponse(op)
}

type Server struct {
	port   uint16
	echo   *echo.Echo
	logger *slog.Logger
}

func (srv *Server) Start(controllers []Controller) error {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.RequestID())

	apiBase := "/api/v1beta"
	api := e.Group(apiBase)
	apiCfg := huma.DefaultConfig("Panoptes", "v1beta")
	hg := humaecho.NewWithGroup(e, api, apiCfg)

	hg.OpenAPI().OnAddOperation = append(hg.OpenAPI().OnAddOperation, configureDefaultResponses)
	// Needed to get the docs displaying properly.
	apiCfg.OpenAPI.Servers = []*huma.Server{
		{
			URL: apiBase,
		},
	}

	for _, c := range controllers {
		c.RegisterRoutes(hg, true)
	}

	srv.echo = e

	srv.logger.Info("starting server", "port", srv.port)

	return e.Start(fmt.Sprintf(":%d", srv.port))
}

func (srv *Server) Shutdown(ctx context.Context) error {
	srv.logger.Info("shutting down gracefully...")
	return srv.echo.Shutdown(ctx)
}

func NewServer(port uint16, logger *slog.Logger) *Server {
	return &Server{
		port:   port,
		logger: logger,
	}
}
