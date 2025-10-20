package users

type githubOauthClient interface {
	GetToken(code string) (string, error)
}

type UserHandlers struct {
	githubOauthClient githubOauthClient
}

func NewUserHandlers(githubOauthClient githubOauthClient) *UserHandlers {
	return &UserHandlers{
		githubOauthClient: githubOauthClient,
	}
}
