# xxdk Restlike Single Use Client Example

This mini-repository contains the example logic for running a basic REST-like 
single use client. This is provided by the xx network team as a springboard to 
help consumers better understand our API and how it functions. 

`main.go` contains the crux of the logic. We avoid complicating our example by
avoiding the usage of CLI flags for basic variables you may change in the code.
This file initiates an xxdk E2E client. With that client established, a 
REST-like client is built on top. Using a precanned contact object created
in `restSingleUseServer` this REST-like client contacts the server with a simple
request.

`utils.go` contains utility functions for running the program. In this case,
we provide a tool initializing a log.

## Build Instructions

In these instructions we will go over building a REST-like client using our 
example. In order to build a client which successfully sends a request and 
receives a response, we must first go over how to build and run a REST-like
single use server.

### Building a Server

In order to run a server, the following commands may be run:

```bash
cd restSingleUseServer/
go build -o server .
./server 
```

This will initialize the server. You may verify its functionality by checking
the `server.log` file. It is a long-running process which may be 
stopped by a user inputted kill signal. This will create a file 
`restSingleUseServer.xxc`, which is the contact file for the server.
A REST-like client may parse this file in order to send a request to this 
server.  

### Building a Client

Please follow the steps above before continuing to these instructions.
In order to run the client, you must first move the aforementioned 
`restSingleUseServer.xxc` file to the path where you will run the client.

```bash
cd restSingleUseServer/
cp restSingleUseServer.xxc /path/to/restSingleUseClient
```

Once the contact object is local to the client, you may build and run
the client:

```bash
cd restSingleUseClient/
go build -o client .
./client 
```

Once the REST-like client has set up and sent its request, you can verify
by checking the server's log for this string `Request received:`

```bash
grep "Request received"  restSingleUseServer/server.log 
INFO 2022/07/07 10:55:57.623516 Request received: headers:{headers:"This is a header"}  method:1  uri:"handleClient"
INFO 2022/07/07 10:56:21.181945 Request received: headers:{headers:"This is a header"}  method:1  uri:"handleClient"
```

By default, the client sends two requests, synchronous and asynchronous. Both 
requests should be received by the server in order to accomplish a successful
client-server request. 

In order to verify the response, look at the client log for the string 
`Response: `:

```bash
 grep "Response: " restSingleUseClient/client.log 
INFO 2022/07/07 11:43:42.923030 Response: content:"This is content"  headers:{headers:"this is a response"}
INFO 2022/07/07 11:43:50.376968 Response: content:"This is content"  headers:{headers:"this is a response"}
```

As by default, there are two requests received by the server, the client will 
receive two responses. 

