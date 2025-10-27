package v1beta

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/application/bus"
)
type GithubLoginRequestBody struct {
	Code string `json:"code" doc:"The code from a github oauth redirect." required:"true"`
}

type GithubTokenRequest struct {
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
}

func (c *AuthController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID:   "v1.auth.github.login",
		Method:        http.MethodPost,
		Path:          "/auth/github/login",
		Summary:       "Login or create an account via github Oauth.",
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
func (c *AuthController) GetGithubToken(ctx context.Context, req *GithubTokenRequest) (*GithubLoginResponse, error) {
	// err := bus.Dispatch(c.bus, users.SignInOrRegisterViaGithub{
	// 	Code: req.Body.Code,
	// })

	// if err != nil {
	// 	return nil, err
	// }


	// resp, err := c.uh.SigninOrRegisterViaGithub(
	// 	users.SignInOrRegisterViaGithub{
	// 		Code: req.Code,
	// 	},
	// )

	// if err != nil {
	// 	return nil, err
	// }

	return &GithubLoginResponse{
		Body: GithubLoginResponseBody{
			Data: GithubLoginResponseData{
				Token: req.Body.Code,
			},
		},
	}, nil
}
