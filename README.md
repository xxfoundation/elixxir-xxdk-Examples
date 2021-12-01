# xxDK Examples

This repo contains the source code of sample apps used in the xxDK documentation.

## Sample Messaging App

A simple messaging app used to illustrate how to get started with the xxDK. It covers the entire process of integrating the cMix Client API (xxDK) in an application, registering within the xxDK, setting up a connection with the cMix network, setting up listeners, as well as sending and receiving messages.

__Prerequisites__

- Go: [Download and install](https://go.dev/doc/install)
- An [NDF](https://xxnetwork.wiki/index.php/Network_Definition_File_(NDF))

__Basic Usage__

```
$ git clone https://git.xx.network/elixxir/xxdk-examples.git
$ cd xxdk-examples
$ cd sample-messaging-app
$ go mod tidy
$ go run main.go
```