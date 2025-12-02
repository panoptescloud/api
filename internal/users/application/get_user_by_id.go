package users

import (
	"github.com/panoptescloud/api/internal/users/domain"
)

type GetUserByID struct {
	ID domain.UserID
}

func (cmd GetUserByID) GetName() string {
	return "users.query.check_user_exists"
}

func (uh *UserQueryHandler) GetUserByID(dto GetUserByID) (*domain.User, error) {
	return uh.userRepo.ByID(dto.ID)
}
