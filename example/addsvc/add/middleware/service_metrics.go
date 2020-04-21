package middleware

import (
	service "addsvc/add"
	genService "addsvc/add/gen/service"
	"context"
	"github.com/go-kit/kit/metrics"
)

// InstrumentingMiddleware returns a service middleware that instruments
// the number of integers summed and characters concatenated over the lifetime of
// the service.
func InstrumentingMiddleware(ints, chars metrics.Counter) genService.Middleware {
	return func(next service.Service) service.Service {
		return instrumentingMiddleware{
			ints:  ints,
			chars: chars,
			next:  next,
		}
	}
}

type instrumentingMiddleware struct {
	ints  metrics.Counter
	chars metrics.Counter
	next  service.Service
}

func (mw instrumentingMiddleware) Sum(ctx context.Context,req service.SumRequest) (*service.SumResponse, error) {
	res, err := mw.next.Sum(ctx,req)
	if res != nil {
		mw.ints.Add(float64(res.V))
	}
	return res, err
}

func (mw instrumentingMiddleware) Concat(ctx context.Context, req service.ConcatRequest) (*service.ConcatResponse, error) {
	res, err := mw.next.Concat(ctx,req)
	if res != nil {
		mw.chars.Add(float64(len(res.V)))
	}
	return res, err
}

