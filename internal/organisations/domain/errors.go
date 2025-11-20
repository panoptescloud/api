package domain

type ErrNameAlreadyInUse struct{}

func (e ErrNameAlreadyInUse) Error() string {
	return "organisation name already in use"
}
