package validation

import "fmt"

type FieldError struct {
	Key    string
	Errors []string
}

func (fE FieldError) PrefixAll(prefix string) FieldError {
	prefixed := make([]string, len(fE.Errors))

	for i, msg := range fE.Errors {
		prefixed[i] = fmt.Sprintf("%s%s", prefix, msg)
	}

	return FieldError{
		Key:    fE.Key,
		Errors: prefixed,
	}
}

type ValidationError struct {
	Errs []FieldError
}

func (vE ValidationError) PrefixAll(prefix string) ValidationError {
	prefixed := make([]FieldError, len(vE.Errs))

	for i, err := range vE.Errs {
		prefixed[i] = err.PrefixAll(prefix)
	}

	return ValidationError{
		Errs: prefixed,
	}
}

func (err ValidationError) Error() string {
	return "invalid data"
}
