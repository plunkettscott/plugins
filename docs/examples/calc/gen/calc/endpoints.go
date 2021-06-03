// Code generated by goa v3.4.3, DO NOT EDIT.
//
// calc endpoints
//
// Command:
// $ goa gen goa.design/plugins/v3/docs/examples/calc/design -o
// $(GOPATH)/src/goa.design/plugins/docs/examples/calc

package calc

import (
	"context"

	goa "goa.design/goa/v3/pkg"
)

// Endpoints wraps the "calc" service endpoints.
type Endpoints struct {
	Add goa.Endpoint
}

// AddEndpointInput holds both the payload and the server stream of the "add"
// method.
type AddEndpointInput struct {
	// Payload is the method payload.
	Payload *AddPayload
	// Stream is the server stream used by the "add" method to send data.
	Stream AddServerStream
}

// NewEndpoints wraps the methods of the "calc" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		Add: NewAddEndpoint(s),
	}
}

// Use applies the given middleware to all the "calc" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.Add = m(e.Add)
}

// NewAddEndpoint returns an endpoint function that calls the method "add" of
// service "calc".
func NewAddEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		ep := req.(*AddEndpointInput)
		return nil, s.Add(ctx, ep.Payload, ep.Stream)
	}
}
