package main

import (
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/connect"
	"gitlab.com/elixxir/client/restlike"
	restConnect "gitlab.com/elixxir/client/restlike/connect"
	"gitlab.com/elixxir/client/xxdk"
	"io/fs"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Logging
	initLog(1, "server.log")

	// Create a new client object----------------------------------------------
	// NOTE: For some (or all) of these parameters, you may want to use a
	// configuration tool of some kind

	// Set the output contact file path
	contactFilePath := "restConnectServer.xxc"

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

	// Parameters for e2e client & restlike server
	e2eParams := xxdk.GetDefaultE2EParams()
	connParams := connect.DefaultConnectionListParams()

	// Set the restlike parameters
	exampleURI := restlike.URI("handleClient")
	exampleMethod := restlike.Get

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
	net, err := xxdk.LoadCmix(statePath, []byte(statePass),
		xxdk.GetDefaultCMixParams())
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

	// Save contact file-------------------------------------------------------

	// Save the contact file so that client can connect to this server
	writeContact(contactFilePath, identity.GetContact())

	// Start rest-like connect server------------------------------------------
	restlikeServer, err := restConnect.NewServer(identity, net,
		e2eParams, connParams)
	if err != nil {
		jww.FATAL.Panicf("Unable to start restlike connect server: %+v", err)
	}
	jww.INFO.Printf("Initialized restlike connect server")

	// Implement restlike endpoint---------------------------------------------

	// Add endpoint
	err = restlikeServer.GetEndpoints().Add(exampleURI, exampleMethod, Callback)
	if err != nil {
		jww.FATAL.Panicf("Failed to add endpoint to server: %v", err)
	}
	jww.DEBUG.Printf("Added endpoint for restlike connect server")

	// Start network threads---------------------------------------------------

	// Set networkFollowerTimeout to a value of your choice (seconds)
	networkFollowerTimeout := 5 * time.Second
	err = restlikeServer.ConnectServer.Messenger.
		StartNetworkFollower(networkFollowerTimeout)
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
	// Provide a callback that will be signalled when network
	// health status changes
	restlikeServer.ConnectServer.Messenger.
		GetCmix().AddHealthCallback(
		func(isConnected bool) {
			connected <- isConnected
		})
	// Wait until connected or crash on timeout
	waitUntilConnected(connected)

	// Keep app running to receive messages------------------------------------

	// Wait until the user terminates the program
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	jww.DEBUG.Printf("Waiting for SIGTERM signal to close process")
	<-c

	err = restlikeServer.ConnectServer.Messenger.
		StopNetworkFollower()
	if err != nil {
		jww.ERROR.Printf("Failed to stop network follower: %+v", err)
	} else {
		jww.INFO.Printf("Stopped network follower.")
	}

	// Close server on function exit
	restlikeServer.Close()
	jww.INFO.Printf("Closed restlike server")

	os.Exit(0)
}
