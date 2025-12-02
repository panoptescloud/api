package http

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// TODO consts for security schemes
func allowsRefreshCookie(ctx huma.Context) bool {
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

func allowsAuthCookie(ctx huma.Context) bool {
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

func allowsApiKey(ctx huma.Context) bool {
	sec := ctx.Operation().Security

	for _, v := range sec {
		for method := range v {
			if method == "api_key" {
				return true
			}
		}
	}

	return false
}

func (srv *Server) AuthByApiKey(key string, ctx huma.Context, api huma.API, next func(huma.Context)) {
	apiToken := ctx.Header("X-Api-Key")

	if apiToken == "" {
		huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	actor, err := srv.al.ByApiKey(apiToken)

	if err != nil {
		srv.logger.Warn("failed to load actor", "error", err)
		huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if actor == nil {
		huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ctx = huma.WithValue(ctx, "actor", actor)
	ctx.SetHeader("X-Api-Key-Id", actor.ApiKeyID().String())
	ctx.SetHeader("X-Auth-Method", "api_key")

	next(ctx)
}

func (srv *Server) AuthByAuthCookie(cookieValue string, ctx huma.Context, api huma.API, next func(huma.Context)) {
	if cookieValue == "" {
		huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	loadedJWT, err := srv.sessionManager.VerifyJWT(cookieValue)

	if err != nil {
		huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check CSRF for this method
	if !methodsWithoutCSRF[ctx.Method()] {
		xcsrfToken := ctx.Header("X-CSRF-Token")
		csrfCookie, err := huma.ReadCookie(ctx, "csrf_token")

		if err != nil || xcsrfToken != csrfCookie.Value {
			huma.WriteErr(api, ctx, http.StatusForbidden, "Forbidden - invalid CSRF")
			return
		}
	}

	subject, err := loadedJWT.Claims.GetSubject()

	if err != nil {
		srv.logger.Warn("jwt claims missing subject")
		huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	uid, err := uuid.Parse(subject)

	if err != nil {
		srv.logger.Warn("failed to parse subject id as uuid from jwt claim")
		huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}
	actor, err := srv.al.ByUserID(uid)

	if err != nil {
		srv.logger.Warn("failed to load actor")
		huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ctx = huma.WithValue(ctx, "actor", actor)
	ctx.SetHeader("X-User-Id", uid.String())
	ctx.SetHeader("X-Auth-Method", "jwt_cookie")
	next(ctx)
}

func (srv *Server) AuthByRefreshCookie(cookieValue string, ctx huma.Context, api huma.API, next func(huma.Context)) {
	if cookieValue == "" {
		huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	next(ctx)
}

func (srv *Server) NewAuthMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		apiKey := ctx.Header("X-Api-Key")
		authCookie, authCookieErr := huma.ReadCookie(ctx, "auth_token")
		refreshCookie, refreshCookieErr := huma.ReadCookie(ctx, "refresh_token")

		switch true {
		case len(ctx.Operation().Security) == 0:
			next(ctx)
		case allowsApiKey(ctx) && apiKey != "":
			srv.AuthByApiKey(apiKey, ctx, api, next)
		case allowsAuthCookie(ctx) && authCookieErr != nil && authCookie != nil && authCookie.Value != "":
			srv.AuthByAuthCookie(authCookie.Value, ctx, api, next)
		case allowsRefreshCookie(ctx) && refreshCookieErr != nil && refreshCookie != nil && refreshCookie.Value != "":
			srv.AuthByRefreshCookie(refreshCookie.Value, ctx, api, next)
		default:
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		}
	}
}
