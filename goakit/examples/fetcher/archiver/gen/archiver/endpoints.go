// Code generated by goa v3.4.2, DO NOT EDIT.
//
// archiver endpoints
//
// Command:
// $ goa gen goa.design/plugins/v3/goakit/examples/fetcher/archiver/design -o
// $(GOPATH)/src/goa.design/plugins/goakit/examples/fetcher/archiver

package archiver

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints wraps the "archiver" service endpoints.
type Endpoints struct {
	Archive endpoint.Endpoint
	Read    endpoint.Endpoint
}

// NewEndpoints wraps the methods of the "archiver" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		Archive: NewArchiveEndpoint(s),
		Read:    NewReadEndpoint(s),
	}
}

// Use applies the given middleware to all the "archiver" service endpoints.
func (e *Endpoints) Use(m func(endpoint.Endpoint) endpoint.Endpoint) {
	e.Archive = m(e.Archive)
	e.Read = m(e.Read)
}

// NewArchiveEndpoint returns an endpoint function that calls the method
// "archive" of service "archiver".
func NewArchiveEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*ArchivePayload)
		res, err := s.Archive(ctx, p)
		if err != nil {
			return nil, err
		}
		vres := NewViewedArchiveMedia(res, "default")
		return vres, nil
	}
}

// NewReadEndpoint returns an endpoint function that calls the method "read" of
// service "archiver".
func NewReadEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*ReadPayload)
		res, err := s.Read(ctx, p)
		if err != nil {
			return nil, err
		}
		vres := NewViewedArchiveMedia(res, "default")
		return vres, nil
	}
}
