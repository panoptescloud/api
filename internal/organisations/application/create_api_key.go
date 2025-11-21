package application

import (
	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/common/validation"
	"github.com/panoptescloud/api/internal/organisations/domain"
)

type CreateAPIKey struct {
	ID             domain.APIKeyID
	OrganisationID domain.OrganisationID
	Name           string
}

// TODO: ensure ID is not in use
func (cmd CreateAPIKey) Validate(repo organisationValidationRepo) error {
	fieldErrors := []validation.FieldError{}

	if cmd.Name == "" {
		fieldErrors = append(fieldErrors, validation.FieldError{
			Key: "Name",
			Errors: validation.Violations{
				validation.NotEmptyViolation{},
			},
		})
	}

	org, err := repo.ByID(cmd.OrganisationID)

	if err != nil {
		return err
	}

	if org == nil {
		fieldErrors = append(fieldErrors, validation.FieldError{
			Key: "OrganisationID",
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

func (cmd CreateAPIKey) GetName() string {
	return "users.command.create_organisation_api_key"
}

func (h OrganisationCmdHandler) CreateAPIKey(cmd CreateAPIKey) error {
	if err := cmd.Validate(h.organisationRepo); err != nil {
		return err
	}

	name := domain.NewAPIKeyName(cmd.Name)

	key := domain.NewAPIKey(
		cmd.ID,
		cmd.OrganisationID,
		name,
		dto.HashedValue{
			// generate this
			Value: "blah",
		},
	)

	return h.apiKeyRepo.Save(key)
}
