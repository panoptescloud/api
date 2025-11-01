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

type jwtService interface {
	Verify(tokenString string) (*jwt.Token, error)
	Generate(subject string, name string) (string, error)
}

type GithubLoginRequestBody struct {
	Code string `json:"code" doc:"The code from a github oauth redirect." required:"true"`
}

type GithubLoginRequest struct {
	Body GithubLoginRequestBody
}

type GithubLoginResponseData struct {
	Token string `json:"token"`
}

type GithubLoginResponseBody struct {
	Data GithubLoginResponseData `json:"data"`
}

type GithubLoginResponse struct {
	Body GithubLoginResponseBody
}

type AuthController struct {
	bus *bus.Bus
	githubOauthClient githubOauthClient
	logger *slog.Logger
	jwtService jwtService
}

func (c *AuthController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID:   "v1.auth.github.login",
		Method:        http.MethodPost,
		Path:          "/auth/github/login",
		Summary:       "Login or create an account via github Oauth.",
		DefaultStatus: http.StatusOK,
		// TODO: figure out which statuses should return here and implement them
		Metadata: map[string]any{
			operations.OptDisableAuthentication: true,
			operations.OptDisableAllDefaults: true,
		},
	}, ErrorHandler(debugErrorsEnabled, c.LoginWithGithub))
}

func NewAuthController(b *bus.Bus, githubOauthClient githubOauthClient, jwtService jwtService, logger *slog.Logger) *AuthController {
	return &AuthController{
		bus: b,
		githubOauthClient: githubOauthClient,
		logger: logger,
		jwtService: jwtService,
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

func (c *AuthController) generateJWT(user *users.User) (string, error) {
	return "blah", nil
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
func (c *AuthController) LoginWithGithub(ctx context.Context, req *GithubLoginRequest) (*GithubLoginResponse, error) {
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

	jwt, err := c.jwtService.Generate(user.ID().String(), user.Name().String())

	if err != nil {
		return nil, err
	}

	// simply generate token
	return &GithubLoginResponse{
		Body: GithubLoginResponseBody{
			Data: GithubLoginResponseData{
				Token: jwt,
			},
		},
	}, nil
}
