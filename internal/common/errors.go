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

type ErrEmailAlreadyInUse struct{}

func (err ErrEmailAlreadyInUse) Error() string {
	return "the chosen email is already in use"
}

type ErrMustHaveVerfiiedEmailAddress struct{}

func (err ErrMustHaveVerfiiedEmailAddress) Error() string {
	return "the user must have a verified email address in the external oauth provider"
}
