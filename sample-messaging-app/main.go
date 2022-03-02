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
)

func main() {

	// Create a new client object-------------------------------------------------------

	// You would ideally use a configuration tool to acquire these parameters
	statePath := "statePath"
	statePass := "password"
	// The following connects to mainnet. For historical reasons it is called a json file
	// but it is actually a marshalled file with a cryptographic signature attached.
	// This may change in the future.
	ndfURL := "https://elixxir-bins.s3.us-west-1.amazonaws.com/ndf/mainnet.json"
	certificatePath := "mainnet.crt"
	ndfPath := "ndf.json"

	// Create the client if there is no session
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		var ndfJSON []byte
		if ndfPath != "" {
			ndfJSON, err = ioutil.ReadFile(ndfPath)
			if err != nil {
				fmt.Printf("Could not read NDF: %+v", err)
			}
		}
		if ndfJSON == nil {
			cert, err := ioutil.ReadFile(certificatePath)
			if err != nil {
				fmt.Printf("Failed to read certificate: %v", err)
			}

			ndfJSON, err = api.DownloadAndVerifySignedNdfWithUrl(ndfURL, string(cert))
			if err != nil {
				fmt.Printf("Failed to download NDF: %+v", err)
			}
		}
		err = api.NewClient(string(ndfJSON), statePath, []byte(statePass), "")
		if err != nil {
			fmt.Printf("Failed to create new client: %+v", err)
		}
	}

	// Login to your client session-----------------------------------------------------

	// Login with the same sessionPath and sessionPass used to call NewClient()
	// Assumes you have imported "gitlab.com/elixxir/client/interfaces/params"
	client, err := api.Login(statePath, []byte(statePass), params.GetDefaultNetwork())
	if err != nil {
		fmt.Printf("Failed to initialize client: %+v", err)
	}

	// view current user identity--------------------------------------------------------
	user := client.GetUser()
	fmt.Println(user)

	// Register a listener for messages--------------------------------------------------

	// Set up a reception handler
	swboard := client.GetSwitchboard()
	// Note: the receiverChannel needs to be large enough that your reception thread will
	// process the messages. If it is too small, messages can be dropped or important xxDK
	// threads could be blocked.
	receiverChannel := make(chan message.Receive, 10000)
	// Note that the name `listenerID` is arbitrary
	listenerID := swboard.RegisterChannel("DefaultCLIReceiver",
		switchboard.AnyUser(), message.XxMessage, receiverChannel)
	fmt.Printf("Message ListenerID: %v", listenerID)

	// Start network threads------------------------------------------------------------

	networkFollowerTimeout := 5 * time.Second

	// Set networkFollowerTimeout to a value of your choice (seconds)
	err = client.StartNetworkFollower(networkFollowerTimeout)
	if err != nil {
		fmt.Printf("Failed to start network follower: %+v", err)
	}

	waitUntilConnected := func(connected chan bool) {
		// Assumes you have imported the `time` package
		waitTimeout := time.Duration(150)
		timeoutTimer := time.NewTimer(waitTimeout * time.Second)
		isConnected := false
		// Wait until we connect or panic if we cannot by a timeout
		for !isConnected {
			select {
			case isConnected = <-connected:
				fmt.Printf("Network Status: %v\n",
					isConnected)
				break
			case <-timeoutTimer.C:
				fmt.Printf("timeout on connection")
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
	confirmChanRequest := func(requestor contact.Contact) {
		// Check if a channel exists for this recipientID
		recipientID := requestor.ID
		if client.HasAuthenticatedChannel(recipientID) {
			fmt.Printf("Authenticated channel already in place for %s",
				recipientID)
			return
		}
		// GetAuthenticatedChannelRequest returns the contact received in a request if
		// one exists for the given userID.  Returns an error if no contact is found.
		recipientContact, err := client.GetAuthenticatedChannelRequest(recipientID)
		if err == nil {
			fmt.Printf("Accepting existing channel request for %s",
				recipientID)
			// ConfirmAuthenticatedChannel() creates an authenticated channel out of a valid
			// received request and informs the requestor that their request has
			// been confirmed
			roundID, err := client.ConfirmAuthenticatedChannel(recipientContact)
			fmt.Printf("Accepted existing channel request in round %v",
				roundID)
			if err != nil {
				fmt.Printf("%+v", err)
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
	contactData, _ := ioutil.ReadFile("../user1b/user-contact1b.json")
	// Assumes you have imported "gitlab.com/elixxir/crypto/contact"
	// which provides an `Unmarshal` function to convert the byte slice ([]byte) output
	// of `ioutil.ReadFile()` to the `Contact` type expected by `RequestAuthenticatedChannel()`
	recipientContact, _ := contact.Unmarshal(contactData)
	recipientID := recipientContact.ID

	roundID, authReqErr := client.RequestAuthenticatedChannel(recipientContact, me, "Hi! Let's connect!")
	if authReqErr == nil {
		fmt.Printf("Requested auth channel from: %s in round %d",
			recipientID, roundID)
	} else {
		fmt.Printf("%+v", err)
	}

	// Send a message to another user----------------------------------------------------

	// Send safe message with authenticated channel, requires an authenticated channel

	// Test message
	msgBody := "If this message is sent successfully, we'll have established first contact with aliens."

	msg := message.Send{
		Recipient:   recipientID,
		Payload:     []byte(msgBody),
		MessageType: message.XxMessage,
	}
	// Get default network parameters for E2E payloads
	paramsE2E := params.GetDefaultE2E()

	fmt.Printf("Sending to %s: %s\n", recipientID, msgBody)
	roundIDs, _, _, err := client.SendE2E(msg,
		paramsE2E)
	if err != nil {
		fmt.Printf("%+v", err)
	}
	fmt.Printf("Message sent in RoundIDs: %+v\n", roundIDs)

	// Keep app running to receive messages-----------------------------------------------
	for {
		msg := <-receiverChannel
		fmt.Println(string(msg.Payload))
	}
}
