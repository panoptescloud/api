package domain

type ErrNameAlreadyInUse struct{}

func (e ErrNameAlreadyInUse) Error() string {
	return "organisation name already in use"
}

type ErrAPIKeyNameAlreadyInUse struct{}

func (e ErrAPIKeyNameAlreadyInUse) Error() string {
	return "organisation api key - name already in use"
}

type ErrTokenConflict struct{}

func (e ErrTokenConflict) Error() string {
	return "organisation api key - token already in use"
}
