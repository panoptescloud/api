package v1beta

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/panoptescloud/api/internal/api/http/operations"
	"github.com/panoptescloud/api/internal/common"
	"github.com/panoptescloud/api/internal/common/bus"
	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/common/validation"
	usersapp "github.com/panoptescloud/api/internal/users/application"
	usersdomain "github.com/panoptescloud/api/internal/users/domain"
)

type githubOauthClient interface {
	GetToken(code string) (string, error)
	GetProfile(accessToken string) (usersdomain.GithubProfile, error)
}

type sessionManager interface {
	VerifyJWT(tokenString string) (*jwt.Token, error)
	Create(userID usersdomain.UserID) (dto.Session, error)
	Refresh(hashedToken dto.HashedValue) (dto.Session, error)
	DeleteRefreshToken(hashedToken dto.HashedValue) error
}

type GithubLoginRequestBody struct {
	Code string `json:"code" doc:"The code from a github oauth redirect." required:"true"`
}

type GithubLoginRequest struct {
	Body GithubLoginRequestBody
}

type User struct {
	ID string `json:"id"`
}

type SessionResponseData struct {
	CSRFToken string `json:"csrf_token"`
	User      User   `json:"user"`
}

type SessionResponseBody struct {
	Data SessionResponseData `json:"data"`
}

// TODO: this looks awful in the spec viewer, see if we can improve it.
type SessionResponse struct {
	SetCookie []http.Cookie `header:"Set-Cookie" description:"Cookies set by the backend: auth_token (httpOnly, JWT) and csrf_token (JS-readable, for X-CSRF-Token header)"`
	Body      SessionResponseBody
}

type AuthController struct {
	bus               *bus.Bus
	githubOauthClient githubOauthClient
	logger            *slog.Logger
	sessionManager    sessionManager
}

func (c *AuthController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID: "v1.auth.github.login",
		Method:      http.MethodPost,
		Path:        "/auth/github/login",
		Summary:     "Login or create an account via github Oauth.",
		Tags: []string{
			"Authentication",
		},
		DefaultStatus: http.StatusOK,
		// TODO: figure out which statuses should return here and implement them
		Metadata: map[string]any{
			operations.OptDisableDefaultAuthentication: true,
			operations.OptDisableAllDefaultResponses:   true,
		},
	}, ErrorHandler(debugErrorsEnabled, c.LoginWithGithub))

	huma.Register(api, huma.Operation{
		OperationID: "v1.auth.refresh",
		Method:      http.MethodPost,
		Path:        "/auth/tokens/refresh",
		Summary:     "Refresh a JWT token",
		Tags: []string{
			"Authentication",
		},
		DefaultStatus: http.StatusOK,
		// TODO: figure out which statuses should return here and implement them
		Metadata: map[string]any{
			operations.OptDisableAllDefaultResponses:   true,
			operations.OptDisableDefaultAuthentication: true,
		},
		Security: []map[string][]string{
			{
				"refresh": []string{},
			},
		},
	}, ErrorHandler(debugErrorsEnabled, c.Refresh))

	huma.Register(api, huma.Operation{
		OperationID: "v1.auth.logout",
		Method:      http.MethodDelete,
		Path:        "/auth/logout",
		Summary:     "Removes the refresh token.",
		Description: "The auth token may still work for a short while, but we have a short token expiry.",
		Tags: []string{
			"Authentication",
		},
		DefaultStatus: http.StatusOK,
		Metadata: map[string]any{
			operations.OptDisableAllDefaultResponses:   true,
			operations.OptDisableDefaultAuthentication: true,
		},
		Security: []map[string][]string{
			{
				"refresh": []string{},
			},
		},
	}, ErrorHandler(debugErrorsEnabled, c.Logout))
}

func NewAuthController(b *bus.Bus, githubOauthClient githubOauthClient, authTokenManager sessionManager, logger *slog.Logger) *AuthController {
	return &AuthController{
		bus:               b,
		githubOauthClient: githubOauthClient,
		logger:            logger,
		sessionManager:    authTokenManager,
	}
}

func (c *AuthController) createAccountFromGithubProfile(profile usersdomain.GithubProfile) (*usersdomain.User, error) {
	id, err := usersdomain.GenerateUserID()

	if err != nil {
		return nil, err
	}

	dto := usersapp.CreateUser{
		ID:           id,
		Email:        profile.Email,
		Name:         profile.Name,
		GithubNodeID: profile.NodeID,
	}

	c.logger.Debug("creating user", "dto", dto)

	if err := bus.Dispatch(c.bus, dto); err != nil {
		return nil, err
	}

	return bus.RunQuery[usersapp.GetUserByGithubNodeID, *usersdomain.User](c.bus, usersapp.GetUserByGithubNodeID{
		NodeID: profile.NodeID,
	})
}

func buildSessionResponse(session dto.Session) *SessionResponse {
	return &SessionResponse{
		SetCookie: []http.Cookie{
			{
				Name:     "auth_token",
				Value:    session.JWT,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   int(session.JWTExpiresAt.Sub(session.IssuedAt).Seconds()),
			},
			{
				Name:     "csrf_token",
				Value:    session.CSRFToken,
				Path:     "/",
				HttpOnly: false,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   int(session.JWTExpiresAt.Sub(session.IssuedAt).Seconds()),
			},
			{
				Name:     "refresh_token",
				Value:    session.RefreshToken.Token.Value,
				Path:     "/",
				HttpOnly: false,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   int(session.RefreshToken.ExpiresAt.Sub(session.IssuedAt).Seconds()),
			},
		},
		Body: SessionResponseBody{
			Data: SessionResponseData{
				User: User{
					ID: session.UserID,
				},
				CSRFToken: session.CSRFToken,
			},
		},
	}
}

func (c *AuthController) handleUserCreationError(err error) error {
	validationErr, ok := err.(validation.Error)
	if !ok {
		return err
	}

	for _, e := range validationErr.FieldErrors {
		// If any field has a not empty violation, something went wrong our side
		// Log the violations.
		if e.Errors.Has(validation.NotEmptyViolationType) {
			c.logger.Error("empty fields during user creation", "field_errors", validationErr.FieldErrors)
			return common.ErrInternalError{}
		}

		// If the email is invalid it likely means our validation isn't supporting
		// the same rules as a third-party (where we got the email from in the
		// first place). This isn't something the client can fix.
		if e.Key == "Email" && e.Errors.Has(validation.EmailViolationType) {
			c.logger.Error("invalid email during user creation", "field_errors", validationErr.FieldErrors)
			return common.ErrInternalError{}
		}

		// TODO: handle email already in use; return a conflict error
		// needed to tell the consumer more info. Need to also ensure rate limiting
		// or recaptcha etc to prevent email enumeration.
	}

	// Last ditch, something else went wrong
	c.logger.Error("error during user creation", "err", err)
	return common.ErrInternalError{}
}

func (c *AuthController) LoginWithGithub(ctx context.Context, req *GithubLoginRequest) (*SessionResponse, error) {
	token, err := c.githubOauthClient.GetToken(req.Body.Code)

	if err != nil {
		return nil, err
	}

	profile, err := c.githubOauthClient.GetProfile(token)

	if err != nil {
		return nil, err
	}

	user, err := bus.RunQuery[usersapp.GetUserByGithubNodeID, *usersdomain.User](c.bus, usersapp.GetUserByGithubNodeID{
		NodeID: profile.NodeID,
	})

	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = c.createAccountFromGithubProfile(profile)

		if err != nil {
			c.logger.Debug("user creation failed", "error", err)
			return nil, c.handleUserCreationError(err)
		}
	}

	session, err := c.sessionManager.Create(user.ID())

	if err != nil {
		return nil, err
	}

	return buildSessionResponse(session), nil
}

type RefreshRequest struct {
	RefreshToken string `cookie:"refresh_token"`
}

func (c *AuthController) Refresh(ctx context.Context, req *RefreshRequest) (*SessionResponse, error) {
	if req.RefreshToken == "" {
		return nil, common.ErrUnauthorised{}
	}

	session, err := c.sessionManager.Refresh(dto.HashedValue{
		Value: req.RefreshToken,
	})

	if err != nil {
		return nil, err
	}

	return buildSessionResponse(session), nil
}

type LogoutRequest struct {
	RefreshToken string `cookie:"refresh_token"`
}

// TODO: this looks awful in the spec viewer, see if we can improve it.
type LogoutResponse struct {
	SetCookie []http.Cookie `header:"Set-Cookie" description:"Cookies set by the backend: auth_token (httpOnly, JWT) and csrf_token (JS-readable, for X-CSRF-Token header)"`
	Status    int
}

// Logout just deletes any cookies that were created. We don't actually do any
// validation of the refresh token or anything here, as we've set the relevant
// security metadata on the route. In the RegisterRoutes method we require the
// "refresh" security scheme. This ensures that the refresh token existed and
// was valid. This check is handled in the server middleware.
// There is a sanity check that isn't empty, but that shouldn't ever happen, just
// being a little defensive.
//
// See: /internal/api/http/server.go->NewAuthMiddleware
func (c *AuthController) Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	resp := &LogoutResponse{
		Status: 204,
		SetCookie: []http.Cookie{
			{
				Name:     "auth_token",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   -1,
			},
			{
				Name:     "csrf_token",
				Value:    "",
				Path:     "/",
				HttpOnly: false,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   -1,
			},
			{
				Name:     "refresh_token",
				Value:    "",
				Path:     "/",
				HttpOnly: false,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   -1,
			},
		},
	}

	if req.RefreshToken == "" {
		return nil, common.ErrUnauthorised{}
	}

	return resp, c.sessionManager.DeleteRefreshToken(dto.HashedValue{
		Value: req.RefreshToken,
	})
}
