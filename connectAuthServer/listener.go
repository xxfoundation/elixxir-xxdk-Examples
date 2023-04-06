package main

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/e2e/receive"
)

// listener adheres to the receive.Listener interface.
type listener struct {
	name string
}

// Hear will be called whenever a message matching the
// RegisterListener call is received.
//
// User-defined message handling logic goes here.
func (l *listener) Hear(item receive.Message) {
	jww.INFO.Printf("Message received: %v", item)
}

// Name is used for debugging purposes.
func (l *listener) Name() string {
	return l.name
}
