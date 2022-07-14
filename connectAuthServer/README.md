# xxdk Authenticated Connection Server Example

This mini-respository contains the example logic for running a basic connection
server with authentication. This will be referred to as a "secure server" for brevity.
This is provided by the xx network team as a springboard to help consumers better 
understand our API and how it may be used.

[`main.go`](main.go) contains the crux of the logic. We avoid complicating our example by
avoiding the usage of CLI flags for basic variables you may change in the code.
This file initiates an xxdk cMix object. With that client established, a 
secure server is built on top. This program creates contact file
`authConnServer.xxc` which may be used by a client to contact the server.

[`utils.go`](utils.go) contains utility functions for running the program. In this case,
we provide a tool initializing a log. It also contains a utility to write the
contact file to disk.

[`listener.go`](listener.go) contains logic for handling the reception of a message via the
established connection. In this example, it is very basic. We invite consumers
to use this as a basis to implement more complex message listeners.


## Build Instructions

In these instructions we will go over building a secure connection server using our
example. This will not include instructions on running a client which
establishes a connection. That documentation may be found in the [`README.md` for
`connectClient`.](../connectAuthClient/README.md).

In order to run a server, the following commands may be run:

```bash
cd connectAuthServer/
go build -o server .
./server 
```

This will initialize the server. You may verify its functionality by checking
the `server.log` file. It is a long-running process which may be stopped by a
user inputted kill signal. This will create a file `authConnServer.xxc`, which is
the contact file for the server. A connection client may parse this file in
order to establish an authenticated connection with this server.  