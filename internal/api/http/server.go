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
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/panoptescloud/api/internal/api/http/operations"
)

type sessionManager interface {
	VerifyJWT(tokenString string) (*jwt.Token, error)
}

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

var methodsWithoutCSRF = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
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

	if v, ok := op.Metadata[operations.OptDisableAllDefaultResponses]; ok {
		if optAsBool, ok := v.(bool); ok && optAsBool {
			return
		}
	}

	addValidationErrorResponse(op)
	addInternalErrorResponse(op)
	addNotFoundResponse(op)
}

func configureDefaultSecurityRequirements(api *huma.OpenAPI, op *huma.Operation) {
	// If this is set in metadata we don't need any auth on that endpoint.
	if _, ok := op.Metadata[operations.OptDisableDefaultAuthentication]; ok {
		return
	}

	schemeName := "auth"

	if len(op.Security) == 0 {
		op.Security = []map[string][]string{
			{schemeName: []string{}},
		}
		return
	}

	// iterate existing requirements
	for i, req := range op.Security {
		if _, ok := req[schemeName]; !ok {
			op.Security[i][schemeName] = []string{}
		}
	}
}

type Server struct {
	echo           *echo.Echo
	logger         *slog.Logger
	sessionManager sessionManager
}

// TODO: account for multiple calls to initialise
func (srv *Server) Initialise(controllers []Controller) {
	srv.echo = echo.New()
	srv.echo.HideBanner = true
	srv.echo.HidePort = true
	srv.echo.Pre(middleware.RemoveTrailingSlash())
	srv.echo.Use(middleware.RequestID())

	apiBase := "/api/v1beta"
	api := srv.echo.Group(apiBase)
	apiCfg := huma.DefaultConfig("Panoptes", "v1beta")
	apiCfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		// Example Authorization Code flow.
		// "bearer": {
		// 	Type: "oauth2",
		// 	Flows: &huma.OAuthFlows{
		// 		AuthorizationCode: &huma.OAuthFlow{
		// 			AuthorizationURL: "https://example.com/oauth/authorize",
		// 			TokenURL:         "https://example.com/oauth/token",
		// 			Scopes: map[string]string{
		// 				"scope1": "Scope 1 description...",
		// 				"scope2": "Scope 2 description...",
		// 			},
		// 		},
		// 	},
		// },

		// Example alternative describing the use of JWTs without documenting how
		// they are issued or which flows might be supported. This is simpler but
		// tells clients less information. Look at the above and see if we can get
		// that working with how we get the code from github on our frontend.
		"auth": {
			Type:        "apiKey",
			Description: "JWT authentication token stored in an HTTP-only cookie named `auth_token`.",
			Name:        "auth_token",
			In:          "cookie",
		},
		"refresh": {
			Type:        "apiKey",
			Description: "A token stored in an HTTP-only cookie named `refresh_token`, used for getting a new auth token and refresh token.",
			Name:        "refresh_token",
			In:          "cookie",
		},
	}
	hg := humaecho.NewWithGroup(srv.echo, api, apiCfg)
	hg.UseMiddleware(NewAuthMiddleware(hg, srv.sessionManager))

	hg.OpenAPI().OnAddOperation = append(
		hg.OpenAPI().OnAddOperation,
		// Note, this should come before default responses, as we may want to use
		// the security requirements to configure extra responses based on whether
		// authentication is required.
		configureDefaultSecurityRequirements,
		configureDefaultResponses,
	)

	// Needed to get the docs displaying properly.
	apiCfg.OpenAPI.Servers = []*huma.Server{
		{
			URL: apiBase,
		},
	}

	for _, c := range controllers {
		c.RegisterRoutes(hg, true)
	}
}

func (srv *Server) GetEcho() *echo.Echo {
	return srv.echo
}

func (srv *Server) Start(port uint16) error {
	srv.logger.Info("starting server", "port", port)

	return srv.echo.Start(fmt.Sprintf(":%d", port))
}

func (srv *Server) Shutdown(ctx context.Context) error {
	srv.logger.Info("shutting down gracefully...")
	return srv.echo.Shutdown(ctx)
}

// TODO consts for security schemes
func requiresRefreshToken(ctx huma.Context) bool {
	sec := ctx.Operation().Security

	for _, v := range sec {
		for method := range v {
			if method == "refresh" {
				return true
			}
		}
	}

	return false
}

func requiresAuthToken(ctx huma.Context) bool {
	sec := ctx.Operation().Security

	for _, v := range sec {
		for method := range v {
			if method == "auth" {
				return true
			}
		}
	}

	return false
}

func requiresCSRFToken(ctx huma.Context) bool {
	return !methodsWithoutCSRF[ctx.Method()] && requiresAuthToken(ctx)
}

func NewAuthMiddleware(api huma.API, jwtService sessionManager) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		authTokenRequired := requiresAuthToken(ctx)
		refreshTokenRequired := requiresRefreshToken(ctx)
		csrfTokenRequired := requiresCSRFToken(ctx)

		if !authTokenRequired && !refreshTokenRequired && !csrfTokenRequired {
			next(ctx)
			return
		}

		if authTokenRequired {
			// Verify JWT from httpOnly cookie
			authCookie, err := huma.ReadCookie(ctx, "auth_token")
			if err != nil || authCookie.Value == "" {
				huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
				return
			}
			if _, err := jwtService.VerifyJWT(authCookie.Value); err != nil {
				huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
				return
			}
		}

		if csrfTokenRequired {
			xcsrfToken := ctx.Header("X-CSRF-Token")
			xcsrfCookie, err := huma.ReadCookie(ctx, "csrf_token")
			if err != nil || xcsrfToken != xcsrfCookie.Value {
				huma.WriteErr(api, ctx, http.StatusForbidden, "Forbidden - invalid CSRF")
				return
			}

		}

		if refreshTokenRequired {
			authCookie, err := huma.ReadCookie(ctx, "refresh_token")
			if err != nil || authCookie.Value == "" {
				huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
				return
			}
		}

		next(ctx)
	}
}

func NewServer(sessionManager sessionManager, logger *slog.Logger) *Server {
	return &Server{
		logger:         logger,
		sessionManager: sessionManager,
	}
}
