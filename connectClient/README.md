# xxdk Connect Client Example

This mini-repository contains the example logic for running a basic connection
client. This is provided by the xx network team as a springboard to
help consumers better understand our API and how it may be used.

`main.go` contains the crux of the logic. We avoid complicating our example by
avoiding the usage of CLI flags for basic variables you may change in the code.
This file initiates an xxdk E2E client. With that established, a connection 
client is built on top. Using a precanned contact object created in 
`connectServer` this connection client contacts the server with a simple 
message.

`utils.go` contains utility functions for running the program. In this case,
we provide a tool initializing a log.

`listener.go` contains logic for handling the reception of a message via the
established connection. In this example, it is very basic. We invite consumers
to use this as a basis to implement more complex message listeners.

## Build Instructions

In these instructions we will go over building a connection client using our
example. In order to build a client which successfully sends a message through
the connection, we must first go over how to build and run a connection server.

### Building a Server

In order to run a server, the following commands may be run:

```bash
cd connectServer/
go build -o server .
./server
```

This will initialize the server. You may verify its functionality by checking
the `server.log`file. It is a long-running process which may be
stopped by a user inputted kill signal. This will create a file
`connectServer.xxc`, which is the contat file for the server. A connection
client may parse this file in order to send a request to this server.

### Building a Client

Please follow the steps above before continuing to these instructions.
In order to run the client, you must first move the aforementioned
`connectServer.xxc` file to the path where you will run the client.

```bash
cd connectServer/
cp connectServer.xxc /path/to/connectClient
```

Once the contact object is local to the client, you may build and run
the client:

```bash
cd connectClient/
go build -o client .
./client 
```

This is a long-running process which may be stopped by a user inputted kill
signal. We recommend allowing the process to run for a long enough time to
complete its requests to the server and receive the server's responses. We go
into detail on what this entails below.

Once the connection client has set up and established its connection with the
server, you can verify by checking the server's log for the string 
`Message received`.

```bash
 grep "Message received" server.log 
INFO 2022/07/07 12:59:12.088046 Message received: {XxMessage WjdMwCH+... [73 102 32 116 104 105 115 32 109 101 115 115 97 103 101 32 105 115 32 115 101 110 116 32 115 117 99 99 101 115 115 102 117 108 108 121 44 32 119 101 39 108 108 32 104 97 118 101 32 101 115 116 97 98 108 105 115 104 101 100 32 99 111 110 116 97 99 116 32 119 105 116 104 32 116 104 101 32 115 101 114 118 101 114 46] kuycotVTjefJ4nZWJ+Ksg9/jviANn6suteW6HPmXroID l74No/qjr/8Q74mA9VadudforXet8OykqSvPIEFAeUQD [0 0 0 0 0 2 245 150] 2022-07-07 12:59:07.078570118 -0700 PDT true {58339144 QUEUED 0xc001e12780 map[PENDING:1969-12-31 16:00:01.65722394 -0800 PST PRECOMPUTING:2022-07-07 12:59:00.644730058 -0700 PDT STANDBY:2022-07-07 12:59:07.062879269 -0700 PDT QUEUED:2022-07-07 12:59:10.062881354 -0700 PDT] [] 1000 18 187058678 ID:58339144  UpdateID:187058678  State:3  BatchSize:1000  Topology:"3\xdd\xc9;\xce\xc5\xf0\xff&\x8c\xf1\x7f\nf\xa8K\x17\xb6\xd1\x0b|a\t[\x14\x8e\xde\xd1qϊB\x02"  Topology:"\xf5\\\x94MB\x19ڣq݃\xbee\x99\xbfF\xb5\xa9\xf3k\x0e8 gl\xf5:d\x11\xab\x89\x17\x02"  Topology:"\x01\xc1\xf6Gi\x972p\xa9\x96\xb4\x12\x0f1\x1c\xebw\xef\xca\xed\"F\xa7w\xe2\n\xbb8\xcbd\x05=\x02"  Topology:"\xd5\xc3\xd00\xa3a;RqDs\xf0\xda<\xa3)$y\xef\xc1\xa0\x12_k?\x00\rIebL\xfe\x02"  Topology:"vQ\xcd\t\xaf\x91ڤ\x86\x8ecl\x84\xb1\x95\x1e\x8f+ږQ\\ﷀ]7\x89\x08\x02"  Timestamps:1657223940  Timestamps:1657223940644730058  Timestamps:1657223947062879269  Timestamps:1657223950062881354  Timestamps:0  Timestamps:0  Timestamps:0  ResourceQueueTimeoutMillis:3906340864  AddressSpaceSize:18  EccSignature:{Nonce:"\xb2y\xccf\x86E\xe0NR\xd2J3|\xb8d\xfe\xb3\xa8\xad\xa2\x92\xe0\xe4\x0bZ\x07\xbeٓ\xb4z\xf2"  Signature:"\xe1\xc9 \x92_\xfe\x9d\x7f\x18\xb920C \xa6\xd1\xe9U\xbb\x93o\x9b\x1bp<Y\xb1\x9f\xb7O\x012^^\x9doa\x06P\x83\xfes\xbf\xe1\xaeL\xb0+\\\xdc\x12r4)\xdas49\xf6=\xd2\x13\xa0\x07"}}}
```

By default, the client sends a single message to the server, with the server
registering a `receive.Listener` which listens to messages of type 
`catalog.NoType` from the client. The server does not send a message back in 
this example. 

