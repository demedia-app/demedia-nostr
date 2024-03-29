<br/>

<p align="center">
  <img src="https://github.com/demedia-app/demedia-app/blob/main/assets/images/Demedia-New-Logo.png?raw=true" width="250" alt="DeMedia Logo"/>
</p>

<h3 align="center">DeMedia</h3>

<p align="center">
   The decentralized social media protocol.
  <br>
  <a href="https://github.com/demedia-app/demedia-paper/blob/main/DeMedia%20Decentralization%20of%20Social%20Media%20-31-1570965830.pdf"><strong>Paper »</strong></a>
</p>

<br/>

DeMedia is a decentralized protocol to facilitate the development of decentralized social media platforms.
The main goal of DeMedia is to provide direct control over the user data to the users themselves.

## How it works

<p align="center">
   <a href='https://postimg.cc/Yh7L49jv' target='_blank'><img src='https://i.postimg.cc/rF01PRkN/image.png' border='0' alt='image' width="650" /></a>
</p>

DeMedia has two main components:

- Hub: Hub provides a social media platform to the users. Peers are connected to the hub and hub fetches and persists the data from the peers.
- Peer: Database is connected to the peer and the peer is connected to the hub.

We implemented a new RPC protocol using [go-libp2p](https://github.com/libp2p/go-libp2p) for communication between hub and peer. 
[Branle](https://github.com/demedia-app/demedia-branle) has a feature to upload an audio file with a message. When a user uploads an audio file, Branle sends the audio file to the hub and hub pins the audio file to the IPFS.
Then hub sends the IPFS hash to the peer and peer persists the IPFS hash to the database.
The user has to set hex for the peer which is used to get static address for the peer. 
Both the hub and peer have been integrated with [OpenTelemetry](https://opentelemetry.io/) for tracing.

## How to run

You have to run a hub and peers on two or more terminals.

### Hub

Open a terminal, then set the environment variables and run with the following commands:

**Environment variables**

```shell
export GIN_MODE=release
export POSTGRESQL_DATABASE=postgres://database_username:database_password@database_host:database_port/database_name?sslmode=disable
export WEB_PORT=3030
export P2P_PORT=10880
export RELAY_PORT=7448
export OTEL_EXPORTER_JAEGER_ENDPOINT=http://address:port/api/traces
export INFURA_PROJECT_ID=infura_project_id
export INFURA_PROJECT_SECRET=infura_project_secret
export IPFS_NODE=https://ipfs.infura.io:5001
export SERVICE_NAME=hub
```

```shell
cd hub
```

```shell
go run main.go
```

### Peer

Open a terminal, then set the environment variables and run with the following commands:

**Environment variables**

```shell
export GIN_MODE=release
export HEX=hex
export HUB=/ip4/hub_ip/tcp/10880/p2p/hub_peer_id
export POSTGRESQL_DATABASE=postgres://database_username:database_password@database_host:database_port/database_name?sslmode=disable
export WEB_PORT=4000
export P2P_PORT=10885
export PORT=7445
export OTEL_EXPORTER_JAEGER_ENDPOINT=http://address:port/api/traces
export SERVICE_NAME=peer-1
```

```shell
cd peer
```

```shell
go run main.go
```
