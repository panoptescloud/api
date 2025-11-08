package users

import (
	"github.com/panoptescloud/api/internal/common/validation"
	"github.com/panoptescloud/api/internal/users/domain"
)

type CreateUser struct {
	ID           domain.UserID
	Name         string
	Email        string
	GithubNodeID string
}

func (cmd CreateUser) Validate(repo userValidationRepo) error {
	fieldErrors := []validation.FieldError{}

	if cmd.Name == "" {
		fieldErrors = append(fieldErrors, validation.FieldError{
			Key: "Name",
			Errors: validation.Violations{
				validation.NotEmptyViolation{},
			},
		})
	}

	if cmd.Email == "" {
		fieldErrors = append(fieldErrors, validation.FieldError{
			Key: "Email",
			Errors: validation.Violations{
				validation.NotEmptyViolation{},
			},
		})
	}

	if cmd.GithubNodeID == "" {
		fieldErrors = append(fieldErrors, validation.FieldError{
			Key: "GithubNodeID",
			Errors: validation.Violations{
				validation.NotEmptyViolation{},
			},
		})
	}

	if len(fieldErrors) > 0 {
		return validation.Error{
			FieldErrors: fieldErrors,
		}
	}

	return nil
}

func (cmd CreateUser) GetName() string {
	return "users.command.create_user"
}

func (uh *UserCmdHandler) CreateUser(dto CreateUser) error {
	if err := dto.Validate(uh.userRepo); err != nil {
		return err
	}

	email, err := domain.NewEmail(dto.Email)
	if err != nil {
		return err
	}

	githubIdentity := domain.NewGithubIdentity(
		domain.NewGithubUserNodeId(dto.GithubNodeID),
	)

	user := domain.NewUser(
		dto.ID,
		domain.NewName(dto.Name),
		email,
		githubIdentity,
	)

	return uh.userRepo.Save(user)
}
