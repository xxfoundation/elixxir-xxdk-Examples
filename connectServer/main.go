// Starting connection server
package main

import (
	"errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/catalog"
	"gitlab.com/elixxir/client/xxdk"
	"io/fs"
	"io/ioutil"
	"os"
	"time"

	"gitlab.com/elixxir/client/connect"
)

func main() {
	// Logging
	initLog(1, "client.log")

	// Create a new client object-------------------------------------------------------

	// Set the output contact file path
	contactFilePath := "connectServer.xxc"

	// You would ideally use a configuration tool to acquire these parameters
	statePath := "statePath"
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

	// Load client state and identity-----------------------------------------------------

	// Load with the same sessionPath and sessionPass used to call NewClient()
	baseClient, err := xxdk.LoadCmix(statePath, []byte(statePass), xxdk.GetDefaultCMixParams())
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

	// Save contact file----------------------------------------------------------------

	// Save the contact file so that client can connect to this server
	writeContact(contactFilePath, identity.GetContact())

	// Handle incoming connections------------------------------------------------------

	// Create callback for incoming connections
	cb := func(connection connect.Connection) {
		// Listen for all types of messages using catalog.NoType
		// User-defined behavior for message reception goes in the listener
		_, err = connection.RegisterListener(
			catalog.NoType, &listener{"connection server listener"})
		if err != nil {
			jww.FATAL.Panicf("Failed to register listener: %+v", err)
		}
	}

	// Start connection server----------------------------------------------------------

	// Start the connection server, which will allow clients to start connections with you
	e2eParams := xxdk.GetDefaultE2EParams()
	connectionListParams := connect.DefaultConnectionListParams()
	connectServer, err := connect.StartServer(
		identity, cb, baseClient, e2eParams, connectionListParams)
	if err != nil {
		jww.FATAL.Panicf("Unable to start connection server: %+v", err)
	}

	// Start network threads------------------------------------------------------------

	// Set networkFollowerTimeout to a value of your choice (seconds)
	networkFollowerTimeout := 5 * time.Second
	err = connectServer.E2e.StartNetworkFollower(networkFollowerTimeout)
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
	connectServer.E2e.GetCmix().AddHealthCallback(
		func(isConnected bool) {
			connected <- isConnected
		})
	// Wait until connected or crash on timeout
	waitUntilConnected(connected)

	// Keep app running to receive messages-----------------------------------------------

	select {}
}
