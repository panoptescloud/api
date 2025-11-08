package v1beta

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/common"
	"github.com/panoptescloud/api/internal/common/validation"
)

func ErrorHandler[Req any, Resp any](debugErrors bool, handler func(context.Context, *Req) (*Resp, error)) func(ctx context.Context, req *Req) (*Resp, error) {
	return func(ctx context.Context, req *Req) (*Resp, error) {
		resp, err := handler(ctx, req)

		if err == nil {
			return resp, nil
		}

		switch e := err.(type) {
		case validation.Error:
			return resp, buildValidationError(e)

		case common.ErrUnauthorised:
			return resp, buildUnauthorisedError(e.Message)

		default:
			// TODO: outside dev, return a generic errors message instead of the
			// actual error as it appears in the response
			return resp, err
		}
	}
}

func buildUnauthorisedError(msg string) huma.StatusError {
	return huma.Error401Unauthorized(msg)
}

func buildValidationError(err validation.Error) *huma.ErrorModel {
	errModel := &huma.ErrorModel{
		Title:  "Invalid data",
		Status: http.StatusUnprocessableEntity,
	}

	for _, fldError := range err.FieldErrors {
		for _, v := range fldError.Errors {
			errModel.Add(&huma.ErrorDetail{
				// TODO: Need to documentthis properly in the API
				// Taking an arguably weird tactic of just returning an "error code"
				// and leaving the consumer to map the error message. Primarily
				// this saves me having to handle any kind of translation.
				Message:  string(v.Code()),
				Location: fldError.Key,
			})
		}
	}
	return errModel
}
