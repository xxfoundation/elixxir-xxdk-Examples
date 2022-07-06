package main

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/restlike"
)

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

// Callback will be called whenever a message is received by the server
//  with a matching restlike.Method and restlike.URI.
//
// If no endpoint exists
// 	the lower level of the restlike package returns an error
//  to the requester.
// User-defined message handling logic goes here.
func (e *endpoint) Callback(
	request *restlike.Message) (response *restlike.Message) {
	jww.INFO.Printf("Request received: %v", request)
	return
}
