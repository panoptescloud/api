// TODO: look at structuring this module better
package validation

type ViolationType string

const NotEmptyViolationType ViolationType = "not_empty"
const EmailViolationType ViolationType = "email"

type Violation interface {
	Code() ViolationType
}

type NotEmptyViolation struct{}

func (v NotEmptyViolation) Code() ViolationType {
	return NotEmptyViolationType
}

type EmailViolation struct{}

func (v EmailViolation) Code() ViolationType {
	return EmailViolationType
}

type Violations []Violation

func (v Violations) Has(t ViolationType) bool {
	for _, i := range v {
		if i.Code() == t {
			return true
		}
	}

	return false
}

type FieldError struct {
	Key    string
	Errors Violations
}

type Error struct {
	FieldErrors []FieldError
}

func (err Error) Error() string {
	return "invalid data"
}
