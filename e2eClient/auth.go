package main

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/cmix/rounds"
	"gitlab.com/elixxir/client/xxdk"
	"gitlab.com/elixxir/crypto/contact"
)

// auth implements the xxdk.AuthCallbacks interface
type auth struct{}

// Request is called when requests are received
// Currently confirms all incoming auth requests
func (a *auth) Request(partner contact.Contact, receptionID receptionID.EphemeralIdentity,
	round rounds.Round, e2e *xxdk.E2e) {
	_, err := e2e.GetAuth().Confirm(partner)
	if err != nil {
		jww.ERROR.Printf("Failed to confirm auth for %s: %+v", partner.ID.String(), err)
	}
}

func (a *auth) Confirm(partner contact.Contact, receptionID receptionID.EphemeralIdentity,
	round rounds.Round, e2e *xxdk.E2e) {
}

func (a *auth) Reset(partner contact.Contact, receptionID receptionID.EphemeralIdentity,
	round rounds.Round, e2e *xxdk.E2e) {
}
