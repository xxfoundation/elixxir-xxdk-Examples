package main

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/restlike"
)

// Callback will be called whenever a message is received by the server
//
//	with a matching restlike.Method and restlike.URI.
//
// If no endpoint exists
//
//		the lower level of the restlike package returns an error
//	 to the requester.
//
// User-defined message handling logic goes here.
func Callback(request *restlike.Message) *restlike.Message {
	jww.INFO.Printf("Request received: %v", request)
	response := &restlike.Message{}
	response.Headers = &restlike.Headers{Headers: []byte("this is a response")}
	response.Content = []byte("This is content")
	return response
}
