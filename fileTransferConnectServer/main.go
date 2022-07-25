// Starting connection server
package main

import (
	"errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/fileTransfer"
	ftConnect "gitlab.com/elixxir/client/fileTransfer/connect"
	"gitlab.com/elixxir/client/xxdk"
	ftCrypto "gitlab.com/elixxir/crypto/fileTransfer"
	"gitlab.com/xx_network/primitives/id"
	"io/fs"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/elixxir/client/connect"
)

func main() {
	// Logging
	initLog(1, "server.log")

	// Create a new client object----------------------------------------------

	// Set the output contact file path
	contactFilePath := "connectServer.xxc"

	// You would ideally use a configuration tool to acquire these parameters
	statePath := "statePath"
	statePass := "password"

	// The following connects to mainnet. For historical reasons it is called a
	// JSON file, but it is actually a marshalled file with a cryptographic
	// signature attached. This may change in the future.
	ndfURL := "https://elixxir-bins.s3.us-west-1.amazonaws.com/ndf/mainnet.json"
	certificatePath := "../mainnet.crt"
	ndfPath := "ndf.json"

	// High level parameters for the network
	e2eParams := xxdk.GetDefaultE2EParams()
	connectionListParams := connect.DefaultConnectionListParams()

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

			ndfJSON, err = xxdk.DownloadAndVerifySignedNdfWithUrl(
				ndfURL, string(cert))
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

	// Load client state and identity------------------------------------------

	// Load with the same sessionPath and sessionPass used to call NewClient()
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

	// Handle incoming connections and create file transfer manager------------

	// Create callback for incoming connections
	cb := func(connection connect.Connection) {

		// Create general file transfer manager
		ftParams := fileTransfer.DefaultParams()
		ftManager, err := fileTransfer.NewManager(ftParams, identity.ID,
			net.GetCmix(), net.GetStorage(), net.GetRng())
		if err != nil {
			jww.FATAL.Panicf("Failed to create file transfer manager: %+v", err)
		}

		err = net.AddService(ftManager.StartProcesses)
		if err != nil {
			jww.FATAL.Panicf("Failed to start file transfer pcoesses: %+v", err)
		}

		// Create ReceiveCallback that is called when a new file transfer is received
		receiveCB := func(tid *ftCrypto.TransferID, fileName, fileType string,
			sender *id.ID, size uint32, preview []byte) {
			// Inform the user that a new file transfer has been received
			jww.INFO.Printf("Received new file %q transfer %s of type %s from %s "+
				"of size %d bytes with preview: %q",
				fileName, tid, fileType, sender, size, preview)

			receivedProgressCB := func(completed bool, received, total uint16,
				rt fileTransfer.ReceivedTransfer, t fileTransfer.FilePartTracker,
				err error) {
				// Show the file transfer progress to the user.
				jww.INFO.Printf("Received progress callback for %q "+
					"{completed: %t, received: %d, total: %d, err: %v}",
					fileName, completed, received, total, err)

				// Once the file is complete, receive the full file.
				if completed {

					fileData, err := ftManager.Receive(tid)
					if err != nil {
						jww.FATAL.Panicf("Failed to receive full file: %+v", err)
					}

					jww.INFO.Printf("Completed receiving file %q: %q",
						fileName, fileData)
				}
			}

			// Register received progress callback
			err = ftManager.RegisterReceivedProgressCallback(
				tid, receivedProgressCB, 0)
			if err != nil {
				jww.FATAL.Panicf(
					"Failed to register received progress callback: %+v", err)
			}
		}

		// Wrap the file transfer in the connection wrapper
		ftConnectParams := ftConnect.DefaultParams()
		_, err = ftConnect.NewWrapper(
			receiveCB, ftConnectParams, ftManager, connection, net.GetCmix())
		if err != nil {
			jww.FATAL.Panicf("Failed to create file transfer manager: %+v", err)
		}

	}

	// Start connection server-------------------------------------------------

	// Start the connection server, which will allow clients to start
	// connections with you
	connectServer, err := connect.StartServer(
		identity, cb, net, e2eParams, connectionListParams)
	if err != nil {
		jww.FATAL.Panicf("Unable to start connection server: %+v", err)
	}

	// Start network threads---------------------------------------------------

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
	// Provide a callback that will be signalled when network health
	// status changes
	connectServer.E2e.GetCmix().AddHealthCallback(
		func(isConnected bool) {
			connected <- isConnected
		})
	// Wait until connected or crash on timeout
	waitUntilConnected(connected)

	// Keep app running to receive messages------------------------------------

	// Wait until the user terminates the program
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	err = connectServer.E2e.StopNetworkFollower()
	if err != nil {
		jww.ERROR.Printf("Failed to stop network follower: %+v", err)
	} else {
		jww.INFO.Printf("Stopped network follower.")
	}

	os.Exit(0)
}
