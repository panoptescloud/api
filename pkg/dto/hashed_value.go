package dto

type HashedValue struct {
	// Original will only be present upon initial creation of this value
	Original *string

	Value string
}
