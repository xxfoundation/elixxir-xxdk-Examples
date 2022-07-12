# xxdk E2E client example

This mini-repository contains example logic for running an e2e client.  
This is provided by the xx network team as a springboard to help consumers 
better understand our API and how it may be used.

`main.go` contains the crux of the logic. We avoid complicating our example by
avoiding the usage of CLI flags for basic variables you may change in the code.
This file initiates an xxdk E2E client, using the authentication callbacks in
`auth.go`. With that established, it registers a generic message listener 
and establishes authentication with a partner.  Finally, it sends a test message 
and listens for incoming messages until stopped by the user.

`utils.go` contains utility functions for running the program. In this case,
we provide a tool initializing a log and one which writes a contact to a file.

`listener.go` contains logic for handling the reception of a message via the
e2e client. In this example, it is very basic. We invite consumers
to use this as a basis to implement more complex message listeners.

## Build Instructions

In these instructions we will go over building a connection client using our
example. In order to build a client which successfully sends a message through
the connection, we must first go over how to build and run a connection server.

### Building a Client
