package v1beta

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/panoptescloud/api/internal/domain/validation"
)

func ErrorHandler[Req any, Resp any](debugErrors bool, handler func(context.Context, *Req) (*Resp, error)) func(ctx context.Context, req *Req) (*Resp, error) {
	return func(ctx context.Context, req *Req) (*Resp, error) {
		resp, err := handler(ctx, req)

		if err == nil {
			return resp, nil
		}

		// TODO: Add error handling logic

		if vErr, ok := err.(validation.ValidationError); ok {
			return resp, buildValidationError(vErr)
		}

		return resp, err
	}
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
