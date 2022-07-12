package main

import (
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/connect"
	"gitlab.com/elixxir/client/e2e"
	"gitlab.com/elixxir/client/restlike"
	restConnect "gitlab.com/elixxir/client/restlike/connect"
	"gitlab.com/elixxir/client/xxdk"
	"gitlab.com/elixxir/crypto/contact"
	"io/fs"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Logging
	initLog(1, "client.log")

	// Create a new client object----------------------------------------------
	// NOTE: For some (or all) of these parameters, you may want to use a
	// configuration tool of some kind

	// Path to the server contact file
	serverContactPath := "restConnectServer.xxc"

	// Set state file parameters
	statePath := "statePath"
	statePass := "password"

	// The following connects to mainnet. For historical reasons
	// it is called a json file but it is actually a marshalled
	// file with a cryptographic signature attached.
	// This may change in the future.
	ndfURL := "https://elixxir-bins.s3.us-west-1.amazonaws.com/ndf/mainnet.json"
	certificatePath := "../mainnet.crt"
	ndfPath := "ndf.json"

	// Set the restlike parameters
	exampleURI := restlike.URI("handleClient")
	exampleMethod := restlike.Get
	exampleContentBytes := []byte("this is some content")
	exampleContent := restlike.Data{}
	copy(exampleContent[:], exampleContentBytes)
	exampleHeaders := &restlike.Headers{
		Headers: []byte("This is a header"),
	}

	// Parameters for e2e client & restlike server
	e2eParams := e2e.GetDefaultParams()

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

			ndfJSON, err = xxdk.DownloadAndVerifySignedNdfWithUrl(ndfURL,
				string(cert))
			if err != nil {
				jww.FATAL.Panicf("Failed to download NDF: %+v", err)
			}
		}

		// Initialize the state using the state file
		err = xxdk.NewCmix(string(ndfJSON), statePath, []byte(statePass), "")
		if err != nil {
			jww.FATAL.Panicf("Failed to initialize state: %+v", err)
		}
	}

	// Login to your client session--------------------------------------------

	// Login with the same sessionPath and sessionPass used to call NewClient()
	baseClient, err := xxdk.LoadCmix(statePath, []byte(statePass),
		xxdk.GetDefaultCMixParams())
	if err != nil {
		jww.FATAL.Panicf("Failed to load state: %+v", err)
	}

	// Get reception identity (automatically created if one does not exist)
	identityStorageKey := "identityStorageKey"
	identity, err := xxdk.LoadReceptionIdentity(identityStorageKey, baseClient)
	if err != nil {
		// If no extant xxdk.ReceptionIdentity, generate and store a new one
		identity, err = xxdk.MakeReceptionIdentity(baseClient)
		if err != nil {
			jww.FATAL.Panicf("Failed to generate reception identity: %+v", err)
		}
		err = xxdk.StoreReceptionIdentity(identityStorageKey, identity, baseClient)
		if err != nil {
			jww.FATAL.Panicf("Failed to store new reception identity: %+v", err)
		}
	}

	// Create an E2E client
	// The 'restlike' package handles AuthCallbacks,
	// xxdk.DefaultAuthCallbacks is fine here
	params := xxdk.GetDefaultE2EParams()
	jww.INFO.Printf("Using E2E parameters: %+v", params)
	e2eClient, err := xxdk.Login(baseClient, xxdk.DefaultAuthCallbacks{},
		identity, params)
	if err != nil {
		jww.FATAL.Panicf("Unable to Login: %+v", err)
	}

	// Start network threads---------------------------------------------------

	// Set networkFollowerTimeout to a value of your choice (seconds)
	networkFollowerTimeout := 5 * time.Second
	err = e2eClient.StartNetworkFollower(networkFollowerTimeout)
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
	// Provide a callback that will be signalled when network health status
	// changes
	e2eClient.GetCmix().AddHealthCallback(
		func(isConnected bool) {
			connected <- isConnected
		})
	// Wait until connected or crash on timeout
	waitUntilConnected(connected)

	// Build contact object----------------------------------------------------

	// Recipient's contact (read from a Client CLI-generated contact file)
	contactData, err := ioutil.ReadFile(serverContactPath)
	if err != nil {
		jww.FATAL.Panicf("Failed to read server contact file: %+v", err)
	}

	// Imported "gitlab.com/elixxir/crypto/contact"
	// which provides an `Unmarshal` function to convert the byte slice ([]byte)
	// output of `ioutil.ReadFile()` to the `Contact` type expected by
	// `RequestAuthenticatedChannel()`
	serverContact, err := contact.Unmarshal(contactData)
	if err != nil {
		jww.FATAL.Panicf("Failed to get contact data: %+v", err)
	}
	jww.INFO.Printf("Recipient contact: %+v", serverContact)

	// Establish connection with the server------------------------------------

	handler, err := connect.Connect(serverContact, e2eClient, params)
	if err != nil {
		jww.FATAL.Panicf("Failed to create connection object: %+v", err)
	}
	jww.INFO.Printf("Connect with %s successfully established!",
		serverContact.ID)

	// Construct request-------------------------------------------------------

	stream := e2eClient.GetRng().GetStream()
	defer stream.Close()

	grp, err := identity.GetGroup()
	if err != nil {
		jww.FATAL.Panicf("Failed to get group from identity: %+v", err)
	}

	request := restConnect.Request{
		Net:    handler,
		Rng:    stream,
		E2eGrp: grp,
	}

	// Send request to the server synchronously--------------------------------
	// This is a synchronous request, meaning it will block until
	// a response is received
	response, err := request.Request(exampleMethod, exampleURI,
		exampleContent, exampleHeaders, e2eParams)
	if err != nil {
		jww.FATAL.Panicf("Failed to call synchronous request "+
			"with server: %+v", err)
	}

	jww.INFO.Printf("Response: %+v", response)

	// Send request to the server asynchronously--------------------------------

	// In order to asynchronously request, a callback is used to handle
	// when the response is received. More complex response handling may be
	// implemented within the `restlike.RequestCallback`.
	responseChan := make(chan *restlike.Message, 1)
	cb := restlike.RequestCallback(func(message *restlike.Message) {
		responseChan <- message
	})

	// Make request
	err = request.AsyncRequest(exampleMethod, exampleURI,
		exampleContent, exampleHeaders, cb, e2eParams)
	if err != nil {
		jww.FATAL.Panicf("Failed to call asynchronous request with server: %+v",
			err)
	}

	response = <-responseChan

	jww.INFO.Printf("Response: %+v", response)

	// Keep app running to receive messages------------------------------------

	// Wait until the user terminates the program
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	jww.DEBUG.Printf("Waiting for SIGTERM signal to close process")
	<-c

	err = e2eClient.StopNetworkFollower()
	if err != nil {
		jww.ERROR.Printf("Failed to stop network follower: %+v", err)
	} else {
		jww.INFO.Printf("Stopped network follower.")
	}

	os.Exit(0)

}
