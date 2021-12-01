package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gitlab.com/elixxir/client/api"
	"gitlab.com/elixxir/client/interfaces/message"
	"gitlab.com/elixxir/client/interfaces/params"
	"gitlab.com/elixxir/client/switchboard"
	"gitlab.com/elixxir/crypto/contact"
	"gitlab.com/xx_network/primitives/id"

	// external
	jww "github.com/spf13/jwalterweatherman" // logging
)

func main() {

	// Create a new client object-------------------------------------------------------

	// You'd ideally use a configuration tool to acquire these parameters
	sessionPath := "sessionPath"
	sessionPass := "sessionPass"

	// Create the client if there's no session
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		// Load NDF (assumes you've saved it to your current working directory)
		// You'd ideally use a configuration tool to acquire this path
		ndfPath := "ndf.json"
		ndfJSON, err := ioutil.ReadFile(ndfPath)
		if err != nil {
			jww.FATAL.Panicf("Failed to read NDF: %+v", err)
		}
		err = api.NewClient(string(ndfJSON), sessionPath, []byte(sessionPass), "")
		if err != nil {
			jww.FATAL.Panicf("Failed to create new client: %+v", err)
		}
	}

	// Login to your client session-----------------------------------------------------

	// Login with the same sessionPath and sessionPass used to call NewClient()
	// Assumes you've imported "gitlab.com/elixxir/client/interfaces/params"
	client, err := api.Login(sessionPath, []byte(sessionPass), params.GetDefaultNetwork())
	if err != nil {
		jww.FATAL.Panicf("Failed to initialize client: %+v", err)
	}

	// view current user identity--------------------------------------------------------
	user := client.GetUser()
	fmt.Println(user)

	// Register a listener for messages--------------------------------------------------

	// Set up a reception handler
	swboard := client.GetSwitchboard()
	receiverChannel := make(chan message.Receive, 10000) // Needs to be large
	// Note that the name `listenerID` is arbitrary
	listenerID := swboard.RegisterChannel("DefaultCLIReceiver",
		switchboard.AnyUser(), message.Text, receiverChannel)
	jww.INFO.Printf("Message ListenerID: %v", listenerID)

	// Start network threads------------------------------------------------------------

  networkFollowerTimeout := 1200

	err = client.StartNetworkFollower(networkFollowerTimeout)
	if err != nil {
		jww.FATAL.Panicf("Failed to start network follower: %+v", err)
	}

	waitUntilConnected := func(connected chan bool) {
		waitTimeout := time.Duration(150)
		timeoutTimer := time.NewTimer(waitTimeout * time.Second)
		isConnected := false
		//Wait until we connect or panic if we can't by a timeout
		for !isConnected {
			select {
			case isConnected = <-connected:
				jww.INFO.Printf("Network Status: %v\n",
					isConnected)
				break
			case <-timeoutTimer.C:
				jww.FATAL.Panic("timeout on connection")
			}
		}
	}

	// Create a tracker channel to be notified of network changes
	connected := make(chan bool, 10)
	// AddChannel() adds a channel to the list of Tracker channels that will be
  // notified of network changes
	client.GetHealth().AddChannel(connected)
	// Wait until connected or crash on timeout
	waitUntilConnected(connected)

	// Register a handler for authenticated channel requests-----------------------------

	// Handler for authenticated channel requests
	confirmChanRequest := func(requestor contact.Contact, message string) {
		// Check if a channel exists for this recipientID
		recipientID := requestor.ID
		if client.HasAuthenticatedChannel(recipientID) {
			jww.INFO.Printf("Authenticated channel already in place for %s",
				recipientID)
			return
		}
		// GetAuthenticatedChannelRequest returns the contact received in a request if
		// one exists for the given userID.  Returns an error if no contact is found.
		recipientContact, err := client.GetAuthenticatedChannelRequest(recipientID)
		if err == nil {
			jww.INFO.Printf("Accepting existing channel request for %s",
				recipientID)
			// ConfirmAuthenticatedChannel() creates an authenticated channel out of a valid
			// received request and informs the requestor that their request has
			// been confirmed
			roundID, err := client.ConfirmAuthenticatedChannel(recipientContact)
			fmt.Println("Accepted existing channel request in round ", roundID)
			jww.INFO.Printf("Accepted existing channel request in round %v",
				roundID)
			if err != nil {
				jww.FATAL.Panicf("%+v", err)
			}
			return
		}
	}

	// Register `confirmChanRequest` as the handler for auth channel requests
	authManager := client.GetAuthRegistrar()
	authManager.AddGeneralRequestCallback(confirmChanRequest)

	// Request auth channels from other users---------------------------------------------

	// Sender's contact for requesting auth channels
	me := client.GetUser().GetContact()
	// Recipient's contact (read from a Client CLI-generated contact file)
	contactData, _ := ioutil.ReadFile("../user2/user-contact.json")
	// Assumes you've imported "gitlab.com/elixxir/crypto/contact" which provides
	// an `Unmarshal` function to convert the byte slice ([]byte) output 
	// of `ioutil.ReadFile()` to the `Contact` type expected by
	// `RequestAuthenticatedChannel()`
	recipientContact, _ := contact.Unmarshal(contactData)
	recipientID := recipientContact.ID

	roundID, authReqErr := client.RequestAuthenticatedChannel(recipientContact, me, "Hi! Let's connect!")
	if authReqErr == nil {
		jww.INFO.Printf("Requested auth channel from: %s in round %d",
			recipientID, roundID)
	} else {
		jww.FATAL.Panicf("%+v", err)
	}

	// Send a message to another user----------------------------------------------------
	msgBody := "If this message is sent successfully, we'll have established first contact with aliens."
	unsafe := client.HasAuthenticatedChannel(recipientID)

	msg := message.Send{
		Recipient:   recipientID,
		Payload:     []byte(msgBody),
		MessageType: message.Text,
	}
	paramsE2E := params.GetDefaultE2E()
	paramsUnsafe := params.GetDefaultUnsafe()

	fmt.Printf("Sending to %s: %s\n", recipientID, msgBody)
	fmt.Println("Sending to: ", recipientID, " , ", msgBody)
	var roundIDs []id.Round
	if unsafe {
		roundIDs, err = client.SendUnsafe(msg,
			paramsUnsafe)
	} else {
		roundIDs, _, _, err = client.SendE2E(msg,
			paramsE2E)
	}
	if err != nil {
		jww.FATAL.Panicf("%+v", err)
	}
	jww.INFO.Printf("RoundIDs: %+v\n", roundIDs)

	// Keep app running to receive messages-----------------------------------------------
	for {
		msg := <-receiverChannel
		fmt.Println(string(msg.Payload))
	}
}