# xxdk Connection Server Example

This mini-repository contains the example logic for running a basic connection
server and receiving a file transfer. This is provided by the xx network team as
a springboard to help consumers better understand our API and how it may be
used.

[`main.go`](main.go) contains the crux of the logic. We avoid complicating our
example by avoiding the usage of CLI flags for basic variables you may change in
the code. This file initiates an xxdk cMix object. With that client established,
a connection server is built on top. When a new connection is received, a file
transfer manager is created that wraps the connection and waits to receive a
file transfer. This program creates contact file  `connectServer.xxc` that may
be used by a client to contact the server.

[`utils.go`](utils.go) contains utility functions for running the program. In this
case, we provide a tool initializing a log. It also contains a utility to write the
contact file to disk.

## Build Instructions

In these instructions we will go over building a connection server using our
example. This will not include instructions on running a client, which 
establishes a connection. That documentation may be found in the [`README.md` for
`connectClient`.](../connectClient/README.md).

In order to run a server, the following commands may be run:

```bash
cd fileTransferConnectServer/
go build -o server .
./server 
```

This will initialize the server. You may verify its functionality by checking
the `server.log` file. It is a long-running process which may be stopped by a
user inputted kill signal. This will create a file `connectServer.xxc`, which is
the contact file for the server. A connection client may parse this file in 
order to establish a connection with this server.  