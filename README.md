# xxDK Examples

This repo contains the source code of sample apps used in the xxDK documentation.

## Sample Messaging App

A simple messaging app used to illustrate how to get started with the xxDK. It covers the entire process of integrating the cMix Client API (xxDK) in an application, registering within the xxDK, setting up a connection with the cMix network, setting up listeners, as well as sending and receiving messages.

__Prerequisites__

- Go: [Download and install](https://go.dev/doc/install)
- An [NDF](https://xxnetwork.wiki/index.php/Network_Definition_File_(NDF))

__Basic Usage__

You can duplicate the app to simulate multiple users and run each instance in different terminals.

```
$ git clone https://git.xx.network/elixxir/xxdk-examples.git
$ cd xxdk-examples
$ cd sample-messaging-app
$ go mod tidy
$ go run main.go
```

__TODO:__

(Currently, one has to poke around and run things manually.)

- Make app run on the command line with user prompts for:
  - sending a message
  - sending/accepting auth request
- Also print received messages
