package users

import "github.com/panoptescloud/api/internal/domain/users"

type CreateUser struct {
	ID           users.UserID
	Name         string
	Email        string
	GithubNodeID string
}

func (cmd CreateUser) GetName() string {
	return "users.command.create_user"
}

func (uh *UserCmdHandler) CreateUser(dto CreateUser) error {
	email, err := users.NewEmail(dto.Email)
	if err != nil {
		return err
	}

	githubIdentity := users.NewGithubIdentity(
		users.NewGithubUserNodeId(dto.GithubNodeID),
	)

	user := users.NewUser(
		dto.ID,
		users.NewName(dto.Name),
		email,
		githubIdentity,
	)

	return uh.userRepo.Save(user)
}
