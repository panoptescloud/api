package http

import "github.com/danielgtaylor/huma/v2"

func (srv *Server) AddOperationIDMiddleware() func(ctx huma.Context, next func(huma.Context)) {
	return func (ctx huma.Context, next func(huma.Context)) {
		ctx.SetHeader("X-Operation-Id", ctx.Operation().OperationID)

		next(ctx)
	}
}