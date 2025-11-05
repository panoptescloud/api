package users

import "github.com/panoptescloud/api/internal/domain/users"

type userQueryRepo interface {
	ByGithubNodeId(id string) (*users.User, error)
	// Save(*users.User) error
}

type userCmdRepo interface {
	Save(*users.User) error
}

type UserQueryHandler struct {
	userRepo userQueryRepo
}

func NewUserQueryHandler(userRepo userQueryRepo) *UserQueryHandler {
	return &UserQueryHandler{
		userRepo: userRepo,
	}
}

type UserCmdHandler struct {
	userRepo userCmdRepo
}

func NewUserCmdHandler(userRepo userCmdRepo) *UserCmdHandler {
	return &UserCmdHandler{
		userRepo: userRepo,
	}
}