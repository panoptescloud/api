// [Users Context]
package users

import (
	"github.com/panoptescloud/api/internal/common/bus"
	"github.com/panoptescloud/api/internal/users/domain"
)

type userQueryRepo interface {
	ByGithubNodeId(id string) (*domain.User, error)
}

type UserQueryHandler struct {
	userRepo userQueryRepo
}

func NewUserQueryHandler(userRepo userQueryRepo) *UserQueryHandler {
	return &UserQueryHandler{
		userRepo: userRepo,
	}
}

type userValidationRepo interface {
	// TODO: add methods required
}

type userCmdRepo interface {
	userValidationRepo
	Save(*domain.User) error
}

type UserCmdHandler struct {
	userRepo userCmdRepo
}

func NewUserCmdHandler(userRepo userCmdRepo) *UserCmdHandler {
	return &UserCmdHandler{
		userRepo: userRepo,
	}
}

type userRepo interface {
	userQueryRepo
	userCmdRepo
}

func RegisterToBus(b *bus.Bus, r userRepo) error {
	cmdHandler := NewUserCmdHandler(r)
	queryHandler := NewUserQueryHandler(r)

	bus.RegisterCommand(b, cmdHandler.CreateUser)
	bus.RegisterQuery(b, queryHandler.GetUserByGithubNodeID)

	return nil
}
