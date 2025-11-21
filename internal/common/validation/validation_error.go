// TODO: look at structuring this module better
package validation

type ViolationType string

const NotEmptyViolationType ViolationType = "not_empty"
const EmailViolationType ViolationType = "email"
const MustExistViolationType ViolationType = "must_exist"
const InvalidUUIDViolationType ViolationType = "invalid_uuid"

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

type MustExistViolation struct{}

func (v MustExistViolation) Code() ViolationType {
	return MustExistViolationType
}

type InvalidUUIDViolation struct{}

func (v InvalidUUIDViolation) Code() ViolationType {
	return InvalidUUIDViolationType
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
