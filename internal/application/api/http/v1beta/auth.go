package v1beta

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/application/bus"
	"github.com/panoptescloud/api/internal/application/users"
)

type GithubTokenRequest struct {
	Code string `query:"code" doc:"The code from a github oauth redirect." required:"true"`
}

type GithubTokenData struct {
	Token string `json:"token"`
}

type GithubTokenBody struct {
	Data GithubTokenData `json:"data"`
}

type GithubTokenResponse struct {
	Body GithubTokenBody
}

type AuthController struct {
	bus *bus.Bus
}

func (c *AuthController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID:   "v1.auth.github.token",
		Method:        http.MethodGet,
		Path:          "/auth/github/token",
		Summary:       "Get an access token via github oauth.",
		DefaultStatus: http.StatusOK,
	}, ErrorHandler(debugErrorsEnabled, c.GetGithubToken))
}

func NewAuthController(b *bus.Bus) *AuthController {
	return &AuthController{
		bus: b,
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
func (c *AuthController) GetGithubToken(ctx context.Context, req *GithubTokenRequest) (*GithubTokenResponse, error) {
	err := bus.Dispatch(c.bus, users.SignInOrRegisterViaGithub{
		Code: req.Code,
	})

	if err != nil {
		return nil, err
	}


	// resp, err := c.uh.SigninOrRegisterViaGithub(
	// 	users.SignInOrRegisterViaGithub{
	// 		Code: req.Code,
	// 	},
	// )

	// if err != nil {
	// 	return nil, err
	// }

	return &GithubTokenResponse{
		Body: GithubTokenBody{
			Data: GithubTokenData{
				Token: "blah",
			},
		},
	}, nil
}
