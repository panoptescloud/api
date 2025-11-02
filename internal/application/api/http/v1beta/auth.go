package v1beta

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/panoptescloud/api/internal/application/api/http/operations"
	"github.com/panoptescloud/api/internal/application/bus"
	appusers "github.com/panoptescloud/api/internal/application/users"
	"github.com/panoptescloud/api/internal/domain/users"
)
type githubOauthClient interface {
	GetToken(code string) (string, error)
	GetProfile(accessToken string) (users.GithubProfile, error)
}

type authTokenManager interface {
	VerifyJWT(tokenString string) (*jwt.Token, error)
	GenerateJWT(userID users.UserID, name string) (string, error)
	RefreshJWT(tokenString string) (string, error)
}

type GithubLoginRequestBody struct {
	Code string `json:"code" doc:"The code from a github oauth redirect." required:"true"`
}

type GithubLoginRequest struct {
	Body GithubLoginRequestBody
}

type JWTResponseData struct {
	Token string `json:"token"`
}

type JWTResponseBody struct {
	Data JWTResponseData `json:"data"`
}

type JWTResponse struct {
	Body JWTResponseBody
}

type AuthController struct {
	bus *bus.Bus
	githubOauthClient githubOauthClient
	logger *slog.Logger
	authTokenManager authTokenManager
}

func (c *AuthController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID:   "v1.auth.github.login",
		Method:        http.MethodPost,
		Path:          "/auth/github/login",
		Summary:       "Login or create an account via github Oauth.",
		Tags: []string{
			"Authentication",
		},
		DefaultStatus: http.StatusOK,
		// TODO: figure out which statuses should return here and implement them
		Metadata: map[string]any{
			operations.OptDisableAuthentication: true,
			operations.OptDisableAllDefaults: true,
		},
	}, ErrorHandler(debugErrorsEnabled, c.LoginWithGithub))

	huma.Register(api, huma.Operation{
		OperationID:   "v1.auth.refresh",
		Method:        http.MethodPost,
		Path:          "/auth/tokens/refresh",
		Summary:       "Refresh a JWT token",
		Tags: []string{
			"Authentication",
		},
		DefaultStatus: http.StatusOK,
		// TODO: figure out which statuses should return here and implement them
		Metadata: map[string]any{
			operations.OptDisableAllDefaults: true,
			operations.OptDisableAuthentication: true,
		},
	}, ErrorHandler(debugErrorsEnabled, c.Refresh))
}

func NewAuthController(b *bus.Bus, githubOauthClient githubOauthClient, authTokenManager authTokenManager, logger *slog.Logger) *AuthController {
	return &AuthController{
		bus: b,
		githubOauthClient: githubOauthClient,
		logger: logger,
		authTokenManager: authTokenManager,
	}
}

func (c *AuthController) createAccountFromGithubProfile(profile users.GithubProfile) (*users.User, error) {
	id, err := users.GenerateUserID()

	if err != nil {
		return nil, err
	}

	dto := appusers.CreateUser{
		ID: id,
		Email: profile.Email,
		Name: profile.Name,
		GithubNodeID: profile.NodeID,
	}

	c.logger.Debug("creating user", "dto", dto)

	err = bus.Dispatch(c.bus, dto)

	if err != nil {
		return nil, err
	}

	return bus.RunQuery[appusers.GetUserByGithubNodeID, *users.User](c.bus, appusers.GetUserByGithubNodeID{
		NodeID: profile.NodeID,
	})
}

/*
Thinking this should:
- call github oauth client to get token (infra)
- retrieve user info from github (infra)
- create user if not exists (application)
	-> use domain to create new user
- generate jwt for user (application)
- return token (application)
*/
func (c *AuthController) LoginWithGithub(ctx context.Context, req *GithubLoginRequest) (*JWTResponse, error) {
	token, err := c.githubOauthClient.GetToken(req.Body.Code)

	if err != nil {
		return nil, err
	}

	profile, err := c.githubOauthClient.GetProfile(token)

	if err != nil {
		return nil, err
	}

	user, err := bus.RunQuery[appusers.GetUserByGithubNodeID, *users.User](c.bus, appusers.GetUserByGithubNodeID{
		NodeID: profile.NodeID,
	})

	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = c.createAccountFromGithubProfile(profile)

		if err != nil {
			return nil, err
		}
	}

	jwt, err := c.authTokenManager.GenerateJWT(user.ID(), user.Name().String())

	if err != nil {
		return nil, err
	}

	// simply generate token
	return &JWTResponse{
		Body: JWTResponseBody{
			Data: JWTResponseData{
				Token: jwt,
			},
		},
	}, nil
}

type RefreshRequestBody struct {
	RefreshToken string `json:"refresh_token" doc:"The token that will be refreshed." required:"true"`
}

type RefreshRequest struct {
	Body RefreshRequestBody
}

func (c *AuthController) Refresh(ctx context.Context, req *RefreshRequest) (*JWTResponse, error) {
	token, err := c.authTokenManager.RefreshJWT(req.Body.RefreshToken)

	if err != nil {
		return nil, err
	}

	return &JWTResponse{
		Body: JWTResponseBody{
			Data: JWTResponseData{
				Token: token,
			},
		},
	}, nil
}
