package application

import (
	"github.com/panoptescloud/api/internal/common/validation"
	"github.com/panoptescloud/api/internal/organisations/domain"
)

type CreateOrganisation struct {
	ID    domain.OrganisationID
	Name  string
	Owner domain.MemberID
}

// TODO: ensure ID is not in use
func (cmd CreateOrganisation) Validate(ub userBridge) error {
	fieldErrors := []validation.FieldError{}

	if cmd.Name == "" {
		fieldErrors = append(fieldErrors, validation.FieldError{
			Key: "Name",
			Errors: validation.Violations{
				validation.NotEmptyViolation{},
			},
		})
	}

	userExists, err := ub.UserExists(cmd.Owner)

	if err != nil {
		return err
	}

	if !userExists {
		fieldErrors = append(fieldErrors, validation.FieldError{
			Key: "Owner",
			Errors: validation.Violations{
				validation.MustExistViolation{},
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

func (cmd CreateOrganisation) GetName() string {
	return "users.command.create_organisation"
}

func (h OrganisationCmdHandler) Create(dto CreateOrganisation) error {
	if err := dto.Validate(h.userBridge); err != nil {
		return err
	}

	name := domain.NewName(dto.Name)
	owner := domain.NewMember(dto.Owner, domain.NewMemberRole("owner"))

	org := domain.NewOrganisation(
		dto.ID,
		name,
		domain.Members{
			owner,
		},
	)

	return h.organisationRepo.Save(org)
}
