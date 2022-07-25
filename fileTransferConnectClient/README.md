# xxdk File Transfer Connect Example

This mini-repository contains example logic for running a basic connection
client and sending a file transfer. This is provided by the xx network team as a
springboard to help consumers better understand our API and how it may be used.

[`main.go`](main.go) contains the crux of the logic. We avoid complicating our
example by avoiding the usage of CLI flags for basic variables you may change in
the code. This file initiates an xxdk E2E object with a connection client built
on top. Then the connection client is wrapped in a file transfer manager that
facilitates the sending of an example file to a precanned contact object created
in `fileTransferConnectServer`.

[`utils.go`](utils.go) contains utility functions for running the program. In
this case, we provide a tool initializing a log.

## Build Instructions

In these instructions we will go over building a connection client using our
example. In order to build a client that successfully sends a message through
the connection, we must first go over how to build and run a connection server.

### Building a Client

E2E clients communicate with each other directly via the xx network.

To build and run a client, execute the following bash commands:
```shell
cd fileTransferConnectClient/
go build -o client .
./client 
```

This will start an e2e client and connection that can be monitored via log
activity in `client.log`. This will read the contact at `connectServer.xxc.xxc`,
attempting to establish a connection and send a test file. This functionality
can be skipped by setting the relevant variable in the code to empty string.  

By default, the client sends a single file to the recipient, with the recipient
registering a `fileTransfer.ReceiveCallback` that listens to messages of type
`catalog.NewFileTransfer` from the connection partner.

Verification that the client is able to send messages may
also be done. This can be done by checking the client's log for the string
`Received progress callback for`.

```shell
grep "Received progress callback for" client.log 
INFO 2022/07/07 12:59:12.088046 Received progress callback for...
```


