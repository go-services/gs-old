package middleware

import (
	service "addsvc/add"
	genService "addsvc/add/gen/service"
	"context"
	"github.com/go-kit/kit/log"
)

func LoggingMiddleware(logger log.Logger) genService.Middleware {
	return func(next service.Service) service.Service {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   service.Service
}

func (mw loggingMiddleware) Sum(ctx context.Context, req service.SumRequest) (res *service.SumResponse,err error) {
	defer func() {
		if res == nil {
			mw.logger.Log("method", "Concat", "a",req.A, "b", req.B, "v", "", "err", err)
			return
		}
		mw.logger.Log("method", "Sum", "a", req.A, "b", req.B, "v", res.V, "err", err)
	}()
	return mw.next.Sum(ctx,req)
}

func (mw loggingMiddleware) Concat(ctx context.Context, req service.ConcatRequest) (res *service.ConcatResponse, err error) {
	defer func() {
		if res == nil {
			mw.logger.Log("method", "Concat", "a",req.A, "b", req.B, "v", "", "err", err)
			return
		}
		mw.logger.Log("method", "Concat", "a",req.A, "b", req.B, "v", res.V, "err", err)
	}()
	return mw.next.Concat(ctx, req)
}

