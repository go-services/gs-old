package add

import (
	"context"
	"errors"
)

var (
	// ErrTwoZeroes is an arbitrary business rule for the Add method.
	ErrTwoZeroes = errors.New("can't sum two zeroes")

	// ErrIntOverflow protects the Add method. We've decided that this error
	// indicates A misbehaving service and should count against e.g. circuit
	// breakers. So, we return it directly in endpoints, to illustrate the
	// difference. In A real service, this probably wouldn't be the case.
	ErrIntOverflow = errors.New("integer overflow")

	// ErrMaxSizeExceeded protects the Concat method.
	ErrMaxSizeExceeded = errors.New("result exceeds maximum size")
)

const (
	intMax = 1<<31 - 1
	intMin = -(intMax + 1)
	maxLen = 10
)


type SumRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}
type SumResponse struct {
	V int `json:"v"`
}

type ConcatRequest struct {
	A string `json:"a"`
	B string `json:"b"`
}
type ConcatResponse struct {
	V string `json:"v"`
}

// @service()
type Service interface {
	// @http(method="post", route="/sum")
	// @grpc()
	Sum(context.Context, SumRequest) (*SumResponse, error)
	// @http(method="post", route="/concat")
	// @grpc()
	Concat(context.Context, ConcatRequest) (*ConcatResponse, error)
}


type addService struct{}

func New() Service {
	return &addService{}
}


func (a addService) Sum(_ context.Context, request SumRequest) (*SumResponse, error) {
	if request.A == 0 && request.B == 0 {
		return nil, ErrTwoZeroes
	}
	if (request.B > 0 && request.A > (intMax-request.B)) || (request.B < 0 && request.A < (intMin-request.B)) {
		return nil, ErrIntOverflow
	}
	return &SumResponse{V: request.A + request.B}, nil
}

func (a addService) Concat(_ context.Context, request ConcatRequest) (*ConcatResponse, error) {
	if len(request.A)+len(request.B) > maxLen {
		return nil, ErrMaxSizeExceeded
	}
	return &ConcatResponse{V: request.A + request.B}, nil
}
