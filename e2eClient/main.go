// Sending Normal messages (Getting Started guide)
package main

import (
	"errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/catalog"
	"gitlab.com/elixxir/client/xxdk"
	"gitlab.com/elixxir/primitives/fact"
	"gitlab.com/xx_network/primitives/id"
	"io/fs"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/elixxir/crypto/contact"
)

func main() {
	// Logging
	initLog(1, "client.log")

	// Create a new client object-------------------------------------------------------

	// Path to the recipient contact file
	recipientContactPath := "myE2eContact.xxc"
	myContactPath := "recipientE2eContact.xxc"

	// You would ideally use a configuration tool to acquire these parameters
	statePath := "statePathRecipient"
	statePass := "password"
	// The following connects to mainnet. For historical reasons it is called a json file
	// but it is actually a marshalled file with a cryptographic signature attached.
	// This may change in the future.
	ndfURL := "https://elixxir-bins.s3.us-west-1.amazonaws.com/ndf/mainnet.json"
	certificatePath := "../mainnet.crt"
	ndfPath := "ndf.json"

	// Check if state exists
	if _, err := os.Stat(statePath); errors.Is(err, fs.ErrNotExist) {

		// Attempt to read the NDF
		var ndfJSON []byte
		ndfJSON, err = ioutil.ReadFile(ndfPath)
		if err != nil {
			jww.INFO.Printf("NDF does not exist: %+v", err)
		}

		// If NDF can't be read, retrieve it remotely
		if ndfJSON == nil {
			cert, err := ioutil.ReadFile(certificatePath)
			if err != nil {
				jww.FATAL.Panicf("Failed to read certificate: %v", err)
			}

			ndfJSON, err = xxdk.DownloadAndVerifySignedNdfWithUrl(ndfURL, string(cert))
			if err != nil {
				jww.FATAL.Panicf("Failed to download NDF: %+v", err)
			}
		}

		// Initialize the state
		err = xxdk.NewCmix(string(ndfJSON), statePath, []byte(statePass), "")
		if err != nil {
			jww.FATAL.Panicf("Failed to initialize state: %+v", err)
		}
	}

	// Login to your client session-----------------------------------------------------

	// Login with the same sessionPath and sessionPass used to call NewClient()
	net, err := xxdk.LoadCmix(statePath, []byte(statePass), xxdk.GetDefaultCMixParams())
	if err != nil {
		jww.FATAL.Panicf("Failed to load state: %+v", err)
	}

	// Get reception identity (automatically created if one does not exist)
	identityStorageKey := "identityStorageKey"
	identity, err := xxdk.LoadReceptionIdentity(identityStorageKey, net)
	if err != nil {
		// If no extant xxdk.ReceptionIdentity, generate and store a new one
		identity, err = xxdk.MakeReceptionIdentity(net)
		if err != nil {
			jww.FATAL.Panicf("Failed to generate reception identity: %+v", err)
		}
		err = xxdk.StoreReceptionIdentity(identityStorageKey, identity, net)
		if err != nil {
			jww.FATAL.Panicf("Failed to store new reception identity: %+v", err)
		}
	}

	writeContact(myContactPath, identity.GetContact())

	// Create an E2E client
	// Pass in auth object which controls auth callbacks for this client
	params := xxdk.GetDefaultE2EParams()
	jww.INFO.Printf("Using E2E parameters: %+v", params)
	confirmChan := make(chan contact.Contact, 5)
	user, err := xxdk.Login(net, &auth{confirmChan: confirmChan}, identity, params)
	if err != nil {
		jww.FATAL.Panicf("Unable to Login: %+v", err)
	}
	e2eClient := user.GetE2E()

	// Start network threads------------------------------------------------------------

	// Set networkFollowerTimeout to a value of your choice (seconds)
	networkFollowerTimeout := 5 * time.Second
	err = user.StartNetworkFollower(networkFollowerTimeout)
	if err != nil {
		jww.FATAL.Panicf("Failed to start network follower: %+v", err)
	}

	// Set up a wait for the network to be connected
	waitUntilConnected := func(connected chan bool) {
		waitTimeout := 30 * time.Second
		timeoutTimer := time.NewTimer(waitTimeout)
		isConnected := false
		// Wait until we connect or panic if we cannot before the timeout
		for !isConnected {
			select {
			case isConnected = <-connected:
				jww.INFO.Printf("Network Status: %v", isConnected)
				break
			case <-timeoutTimer.C:
				jww.FATAL.Panicf("Timeout on starting network follower")
			}
		}
	}

	// Create a tracker channel to be notified of network changes
	connected := make(chan bool, 10)
	// Provide a callback that will be signalled when network health status changes
	user.GetCmix().AddHealthCallback(
		func(isConnected bool) {
			connected <- isConnected
		})
	// Wait until connected or crash on timeout
	waitUntilConnected(connected)

	// Register a listener for messages--------------------------------------------------

	// Listen for all types of messages using catalog.NoType
	// Listen for messages from all users using id.ZeroUser
	// User-defined behavior for message reception goes in the listener
	_ = e2eClient.RegisterListener(&id.ZeroUser, catalog.NoType, listener{name: "e2e Message Listener"})

	// Connect with the recipient--------------------------------------------------

	if recipientContactPath != "" {
		// Wait for 30 seconds to ensure network connectivity
		time.Sleep(30 * time.Second)

		// Recipient's contact (read from a Client CLI-generated contact file)
		contactData, err := ioutil.ReadFile(recipientContactPath)
		if err != nil {
			jww.FATAL.Panicf("Failed to read recipient contact file: %+v", err)
		}

		// Imported "gitlab.com/elixxir/crypto/contact"
		// which provides an `Unmarshal` function to convert the byte slice ([]byte) output
		// of `ioutil.ReadFile()` to the `Contact` type expected by `RequestAuthenticatedChannel()`
		recipientContact, err := contact.Unmarshal(contactData)
		if err != nil {
			jww.FATAL.Panicf("Failed to get contact data: %+v", err)
		}
		jww.INFO.Printf("Recipient contact: %+v", recipientContact)

		// Check that the partner exists, if not send a request
		_, err = e2eClient.GetPartner(recipientContact.ID)
		if err != nil {
			_, err = user.GetAuth().Request(recipientContact, fact.FactList{})
			if err != nil {
				jww.FATAL.Panicf("Failed to send contact request to %s: %+v", recipientContact.ID.String(), err)
			}
			timeout := time.NewTimer(30 * time.Second)
			select {
			case pc := <-confirmChan:
				if !pc.ID.Cmp(recipientContact.ID) {
					jww.FATAL.Panicf("Did not receive confirmation for the requested contact")
				}
				break
			case <-timeout.C:
				jww.FATAL.Panicf("Timed out waiting to receive confirmation of e2e relationship with partner")
			}
		}

		// Send a message to the recipient----------------------------------------------------

		// Test message
		msgBody := "If this message is sent successfully, we'll have established contact with the recipient."
		roundIDs, messageID, timeSent, err := e2eClient.SendE2E(catalog.XxMessage, recipientContact.ID, []byte(msgBody), params.Base)
		if err != nil {
			jww.FATAL.Panicf("Failed to send message: %+v", err)
		}
		jww.INFO.Printf("Message %v sent in RoundIDs: %+v at %v", messageID, roundIDs, timeSent)
	}

	// Keep app running to receive messages-----------------------------------------------

	// Wait until the user terminates the program
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	err = user.StopNetworkFollower()
	if err != nil {
		jww.ERROR.Printf("Failed to stop network follower: %+v", err)
	} else {
		jww.INFO.Printf("Stopped network follower.")
	}

	os.Exit(0)
}
