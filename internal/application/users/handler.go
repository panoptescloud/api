package users

import "github.com/panoptescloud/api/internal/domain/users"

type githubOauthClient interface {
	GetToken(code string) (string, error)
	GetProfile(accessToken string) (users.GithubProfile, error)
}

type userRepo interface {
	ByGithubNodeId(id string) (*users.User, error)
	// Save(*users.User) error
}

type UserHandlers struct {
	githubOauthClient githubOauthClient
	userRepo userRepo
}

func NewUserHandlers(githubOauthClient githubOauthClient, userRepo userRepo) *UserHandlers {
	return &UserHandlers{
		githubOauthClient: githubOauthClient,
		userRepo: userRepo,
	}
}
