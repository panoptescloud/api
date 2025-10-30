package v1beta

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/application/bus"
	appusers "github.com/panoptescloud/api/internal/application/users"
	"github.com/panoptescloud/api/internal/domain/users"
)
type githubOauthClient interface {
	GetToken(code string) (string, error)
	GetProfile(accessToken string) (users.GithubProfile, error)
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
}

func (c *AuthController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID:   "v1.auth.github.login",
		Method:        http.MethodPost,
		Path:          "/auth/github/login",
		Summary:       "Login or create an account via github Oauth.",
		DefaultStatus: http.StatusOK,
	}, ErrorHandler(debugErrorsEnabled, c.LoginWithGithub))
}

func NewAuthController(b *bus.Bus, githubOauthClient githubOauthClient) *AuthController {
	return &AuthController{
		bus: b,
		githubOauthClient: githubOauthClient,
	}
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

	res, err := bus.RunQuery[appusers.GetUserByGithubNodeID, *users.User](c.bus, appusers.GetUserByGithubNodeID{
		NodeID: profile.NodeID,
	})

	if err != nil {
		return nil, err
	}

	if res == nil {
		// Create user, then generate token
		return &GithubLoginResponse{
			Body: GithubLoginResponseBody{
				Data: GithubLoginResponseData{
					Token: "not found",
				},
			},
		}, nil
	}

	// simply generate token
	return &GithubLoginResponse{
		Body: GithubLoginResponseBody{
			Data: GithubLoginResponseData{
				Token: req.Body.Code,
			},
		},
	}, nil
}
