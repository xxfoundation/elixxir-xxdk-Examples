package main

import "gitlab.com/elixxir/client/restlike"

// endpoint implements the restlike.Callback interface.
type endpoint struct {
	uri    restlike.URI
	method restlike.Method
}

// NewEndpoint is a constructor for an endpoint.
func NewEndpoint(uri restlike.URI, method restlike.Method) *endpoint {
	return &endpoint{
		uri:    uri,
		method: method,
	}
}

// Callback adheres to the restlike.Callback interface. It is an example
// of how an endpoint may respond to a request.
func (e *endpoint) Callback(
	request *restlike.Message) (response *restlike.Message) {
	// todo: implment me

	return
}
