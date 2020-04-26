package strings

import (
	"context"
	"errors"
	"strings"
)

type UppercaseRequest struct {
	S string `json:"s"`
}

type UppercaseResponse struct {
	V string `json:"v"`
}
type CountRequest struct {
	S string `json:"s"`
}

type CountResponse struct {
	V int `json:"v"`
}

var ErrEmpty = errors.New("empty string")

// @service()
type Service interface {
	// @http(method="post", route="/uppercase")
	// @grpc()
	Uppercase(context.Context, UppercaseRequest) (*UppercaseResponse, error)
	// @http(method="post", route="/count")
	// @grpc()
	Count(context.Context, CountRequest) (*CountResponse, error)
}

type stringsService struct{}

func New() Service {
	return &stringsService{}
}

func (s stringsService) Uppercase(_ context.Context, req UppercaseRequest) (*UppercaseResponse, error) {
	if req.S == "" {
		return nil, ErrEmpty
	}
	return &UppercaseResponse{V: strings.ToUpper(req.S)}, nil
}

func (s stringsService) Count(_ context.Context, req CountRequest) (*CountResponse, error) {
	return &CountResponse{V: len(req.S)}, nil
}
