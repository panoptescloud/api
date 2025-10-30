package domain

type ErrUnauthorised struct {
	Message string
}

func (err ErrUnauthorised) Error() string {
	if err.Message != "" {
		return err.Message
	}

	return "the current user is not authorized to perform this action"
}