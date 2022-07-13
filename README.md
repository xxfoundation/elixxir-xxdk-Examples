# Xxdk Examples

This repository contains in it examples on how to use the xxdk. 
We refer to the directories in this repo as "mini-repositories". They contain
individual examples on running a product the xx network team provides. Within
each mini-repo there is a comprehensive instruction guide explaining the purpose
of this mini-repo and how to build, run and test this example. Below is a brief 
summary of each mini-repo.

`connectClient` contains an example for implementing a basic connection-style client. 
This is based on the connect/ package found in the xx network's 
[client repo](https://git.xx.network/elixxir/client/-/tree/release/connect).
Building and running this mini-repo will provide you with a client which can send messages 
to a connection-style server.

`connectServer` contains an example for implementing a basic connection-style server.
This is based on the connect/ package found in the xx network's
[client repo](https://git.xx.network/elixxir/client/-/tree/release/connect). 
Building and running this mini-repo will provide you with a long-running server which may receive
connections and messages from a connection-style client.

`restConnectClient` contains an example for implementing a basic REST-like 
connection-style client. This is based on the restlike/connect package
[client repo](https://git.xx.network/elixxir/client/-/tree/release/restlike/connect).
Building and running this mini-repo will provide you with a client which can connect and 
make REST-like requests to a REST-like connection server.

`restConnectServer` contains an example for implementing a basic REST-like
connection-style server. This is based on the restlike/connect package
[client repo](https://git.xx.network/elixxir/client/-/tree/release/restlike/connect).
Building and running this mini-repo will provide you with a long-running which may
receive connections and REST-like requests from a client.

`restSingleUseClient` contains an example for implementing a basic REST-like
single-use client. This is based on the restlike/single package
[client repo](https://git.xx.network/elixxir/client/-/tree/release/restlike/single).
Building and running this mini-repo will provide you with a client which can connect and
make REST-like requests to a REST-like single use server.

`restSingleUseServer` contains an example for implementing a basic REST-like
single-use server. This is based on the restlike/single package
[client repo](https://git.xx.network/elixxir/client/-/tree/release/restlike/single).
Building and running this mini-repo will provide you with a long-running which may
receive connections and REST-like requests from a client.


