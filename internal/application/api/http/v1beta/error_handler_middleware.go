package v1beta

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/domain"
	"github.com/panoptescloud/api/internal/domain/validation"
)

func ErrorHandler[Req any, Resp any](debugErrors bool, handler func(context.Context, *Req) (*Resp, error)) func(ctx context.Context, req *Req) (*Resp, error) {
	return func(ctx context.Context, req *Req) (*Resp, error) {
		resp, err := handler(ctx, req)

		if err == nil {
			return resp, nil
		}

		switch e := err.(type) {
		case validation.ValidationError:
			return resp, buildValidationError(e)

		case domain.ErrUnauthorised:
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

func buildValidationError(err validation.ValidationError) *huma.ErrorModel {
	errModel := &huma.ErrorModel{
		Title:  "Invalid data",
		Status: http.StatusUnprocessableEntity,
	}

	for _, fldError := range err.Errs {
		for _, msg := range fldError.Errors {
			errModel.Add(&huma.ErrorDetail{
				Message:  msg,
				Location: fldError.Key,
			})
		}
	}
	return errModel
}
