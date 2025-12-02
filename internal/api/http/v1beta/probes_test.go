package v1beta_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/panoptescloud/api/internal/api/http/v1beta"
	"github.com/panoptescloud/api/internal/api/http/v1beta/responses"
	"github.com/stretchr/testify/assert"
)

func Test_ProbesController_Startup(t *testing.T) {
	c := v1beta.NewProbesController()

	resp, err := c.Startup(context.TODO(), &v1beta.StartupRequest{})

	assert.Nil(t, err)
	assert.Equal(t, &responses.Empty{
		Status: http.StatusNoContent,
	}, resp)
}
