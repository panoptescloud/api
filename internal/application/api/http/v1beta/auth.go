package v1beta

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/application/users"
)

type userHandlers interface {
	SigninOrRegisterViaGithub(dto users.SignInOrRegisterViaGithubDTO) (users.SignInOrRegisterViaGithubResponse, error)
}

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
	uh userHandlers
}

func (c *AuthController) RegisterRoutes(api huma.API, debugErrorsEnabled bool) {
	huma.Register(api, huma.Operation{
		OperationID:   "v1.auth.github.token",
		Method:        http.MethodGet,
		Path:          "/auth/github/token",
		Summary:       "Get a token from github oauth.",
		DefaultStatus: http.StatusOK,
	}, ErrorHandler(debugErrorsEnabled, c.GetGithubToken))
}

func NewAuthController(uh userHandlers) *AuthController {
	return &AuthController{
		uh: uh,
	}
}

func (c *AuthController) GetGithubToken(ctx context.Context, req *GithubTokenRequest) (*GithubTokenResponse, error) {
	resp, err := c.uh.SigninOrRegisterViaGithub(
		users.SignInOrRegisterViaGithubDTO{
			Code: req.Code,
		},
	)

	if err != nil {
		return nil, err
	}

	return &GithubTokenResponse{
		Body: GithubTokenBody{
			Data: GithubTokenData{
				Token: resp.Token,
			},
		},
	}, nil
}
