package testutil

import (
	"testing"

	"github.com/google/uuid"
)

func NewUuidV7(t *testing.T) uuid.UUID {
	id, err := uuid.NewUUID()

	if err != nil {
		t.Logf("failed to create uuid: %s", err.Error())
		t.FailNow()
	}

	return id
}
