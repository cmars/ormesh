# ormesh - the onion-routed mesh

![mesh bag of onions](onion-mesh.jpg)

ormesh helps you connect services through [Tor](https://www.torproject.org/).

## Why?

Abstract away geography and network topologies.

Disregard container networking, NATs, firewall policies, possibly even traffic
shaping and protocol filtering, if you throw bridges and obfsproxy into the
mix.

Access services running almost anywhere, from just about anywhere else.

## How?

Tor is well-suited to traversing all kinds of networks between services and the
clients that would consume them. Tor provides resilient network infrastructure
with no single point of failure.

Tor hidden services can be deployed in a private, authenticated mode, which
keeps services from being generally accessible.

ormesh helps manage the configuration and auth token exchange necessary to
deploy a private backplane to connect infrastructure.

## What kind of services?

HTTP, email, messaging, sensors & actuators, home automation, and file
synchronization are just some ideas to get you started.

In general, services that require little bandwidth or tolerate latency. With
ormesh, they can be accessed without the hassle of setting up iptables, NAT
port forwarding, VPNs, TLS, and without relying on central rendezvous servers.

## What ormesh isn't

ormesh is not a VPN in the conventional sense.

ormesh is not intended for operating unauthenticated anonymous hidden services.
Anonymity is an interesting side-effect of building on Tor, but it is not a
priority for ormesh, nor it is guaranteed. Users are responsible for evaluating
ormesh (and its Tor configuration) and deciding whether it meets security
requirements and threat models.

Low-latency, high bandwidth applications may not perform well over ormesh's Tor
configuration. Improvements here are possible (by trading anonymity for
improved latency and network throughput) but not yet implemented.

Tor only routes TCP traffic.

# Install

## macOS

[Install Tor Browser](https://www.torproject.org/download/download-easy.html.en). Then:

    brew tap cmars/ormesh
    brew install ormesh

ormesh operates the Tor executable that comes with Tor Browser.

## Windows

[Install Tor Browser](https://www.torproject.org/download/download-easy.html.en).

Like macOS, relies on Tor Browser. The Windows default config expects to find
Tor Browser installed on the current user's Desktop. Not really tested, good
luck.

## Debian & Ubuntu Linux

### curl | bash

Read the script before running if you like. It will install ormesh to /usr/bin,
install Tor from official torproject archives, setcap ormesh to allow
privileged port binding, and install ormesh as a systemd service.

    curl https://git.io/vFN94 -sSfL | bash
    
### Snap packaging

    snap install --edge ormesh

The snap package does not work well for some use cases so it's considered
experimental. I've had trouble installing into containers and binding to
privileged ports.

## Docker

    docker run --name ormesh -d cmars/ormesh:0.2.0

Make it persistent and automatically start up:

    docker run --name ormesh -d \
        -v /srv/ormesh-config:/var/lib/ormesh cmars/ormesh:0.2.0

## Other options

Download an ormesh binary tarball [release](releases) or build from source:

[Install Go](https://golang.org/doc/install).

Download and build ormesh:

    go get -u github.com/cmars/ormesh

# Configuration

## Exporting local services

Export services running locally as Tor hidden services.

```
$ ormesh export add 22
$ ormesh export add 80
$ ormesh status
service:
  export:
    - 127.0.0.1:22
    - 127.0.0.1:80
```

Export services on other hosts.

```
$ ormesh export add 192.168.1.19:8000
```

## Adding clients

Each client gets an auth token string that grants access to the exported
services. Without the auth token, the hidden service is not accessible.

This string should be securely sent to the user of `my-MacBook`:

```
$ ormesh client add my-MacBook
fl3scqcsbitwf7zb.onion x29A3kzv4hrYvBhTkPMV2h
```

## Launch the agent

The agent will operate Tor, implementing the configured export and client
access policies.

```
$ ormesh agent run
```

On Linux, the agent will launch Tor and run it as a subprocess until
interrupted or terminated.

On macOS and Windows, the agent will connect to the Tor process launched with
the Tor Browser and exit after applying changes to the Tor configuration --
unless remote services are imported locally.

# Consuming services

## Add a remote service, with client authentication

On the machine `my-MacBook`, start Tor Browser, and then add a remote using the
onion address and auth token displayed by `client add` above.

```
$ ormesh remote add my-server fl3scqcsbitwf7zb.onion x29A3kzv4hrYvBhTkPMV2h
```

```
$ ormesh remote show my-server
fl3scqcsbitwf7zb.onion
```

## Display an SSH config entry

Display an ssh-config(5) stanza for the remote.

```
$ ormesh remote ssh-config my-server
Host website
  ProxyCommand nc -X 5 -x localhost:9250 %h %p
  Hostname fl3scqcsbitwf7zb.onion
```

## Importing remote services

Set up local port forwarding to remote services with _imports_. The agent will
forward connections to local ports to the corresponding remote service until
the process is interrupted or terminated.

Forward local port 10022 to port 22 on the remote:

```
$ ormesh import add website 22 127.0.0.1:10022
$ ormesh agent run
```

Listen on all addresses to create a public ingress to a remote service. Useful
for circumventing inbound port blocks where the service is running. For
example, you want to physically locate your email server in a mobile camper,
your ISP blocks SMTP inbound, and your IP address changes often. Import your
services from a cloud instance with a public IP and DNS.

```
$ ormesh agent privbind
$ ormesh import add mailinabox 25 0.0.0.0:25
$ ormesh import add mailinabox 587 0.0.0.0:587
$ ormesh agent run
```

# Operating the agent

```
$ ormesh agent run
```

Configuration changes made while the agent is running are applied immediately.

## Setting up systemd

Display a systemd unit file that will run ormesh, from its current installed
binary path.

```
$ ormesh agent systemd
[Unit]
Description=ormesh - onion-routed mesh

[Service]
ExecStart=/path/to/ormesh agent run
Restart=always
User=ubuntu

[Install]
WantedBy=default.target
```

## Docker

The ormesh image supports configuration by environment variables: 

    docker run --name ormesh -d \
        -e 'ORMESH_EXPORTS=80' \
        -e 'ORMESH_CLIENTS=desktop;laptop' cmars/ormesh:0.2.0

will preconfigure ormesh to export 127.0.0.1:80 to clients named "desktop" and
"laptop".

Display the client's onion address & auth cookie by "adding" them again
(`client add` is idempotent):

    docker exec ormesh /ormesh client add desktop

Other configuration commands can be applied with `docker exec` while the
container is running, changes are applied immediately.

