// TODO: Track down usages of thee errors and move them somewhere more appropriate
package common

type ErrInternalError struct {
}

func (e ErrInternalError) Error() string {
	return "an internal error occurred"
}

type ErrUnauthorised struct {
	Message string
}

func (err ErrUnauthorised) Error() string {
	if err.Message != "" {
		return err.Message
	}

	return "the current user is not authorized to perform this action"
}

// TODO: the errors below here should be in the relevant domain, not common!

type ErrEmailAlreadyInUse struct{}

func (err ErrEmailAlreadyInUse) Error() string {
	return "the chosen email is already in use"
}

type ErrOrganisationNameAlreadyInUse struct{}

func (err ErrOrganisationNameAlreadyInUse) Error() string {
	return "the chosen organisation name is already in use"
}

type ErrMustHaveVerfiedEmailAddress struct{}

func (err ErrMustHaveVerfiedEmailAddress) Error() string {
	return "the user must have a verified email address in the external oauth provider"
}

type ErrResourceNotFound struct{}

func (err ErrResourceNotFound) Error() string {
	return "resource not found"
}
