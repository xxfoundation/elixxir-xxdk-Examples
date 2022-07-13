# xxdk Restlike Single Use Server Example

This mini-respository contains the example logic for running a basic REST-like
single use server. This is provided by the xx network team as a springboard 
to help consumers better understand our API and how it may be used.

`main.go` contains the crux of the logic. We avoid complicating our example by
avoiding the usage of CLI flags for basic variables you may change in the code.
This file initiates an xxdk E2E object. With that client established, a 
REST-like server is built on top. This program creates contact file 
`restSingleUseServer.xxc` which may be used by a client to contact the server.

`utils.go` contains utility functions for running the program. In this case,
we provide a tool initializing a log. It also contains a utility to write the
contact file to disk. 

`endpoints.go` contains a simple example of request handling by the server. 
This prints out the request and builds a response to return the to requester.
This endpoint may be modified for more complex request handling, or more 
endpoints with various request handling may be put here.

## Build Instructions

In these instructions we will go over building a REST-like server using our
example. This will not include instructions on running a client which sends
requests. That documentation may be found in the `README.md` for 
`restSingleUseClient`.

In order to run a server, the following commands may be run:

```bash
cd restSingleUseServer/
go build -o server .
./server 
```

This will initialize the server. You may verify its functionality by checking 
the `server.log` file. It is a long-running process which may be stopped by a 
user inputted kill signal. This will create a file `restSingleUseServer.xxc`, 
which is the contact file for the server. A REST-like client may parse this 
file in order to send a request to this server.  
