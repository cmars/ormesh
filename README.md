# ormesh - onion-routed mesh

ormesh helps you build a private mesh of hosts connected through Tor.

Access services running almost anywhere, from almost anywhere else.

Abstract away geography and network topologies.

## Why?

Tor is capable of traversing all kinds of networks between services and the
clients that would consume them. Tor has a resilient infrastructure with no
single point of failure.

Tor hidden services can be deployed in a private, authenticated mode, which
keeps services from being generally accessible over Tor.

ormesh helps manage the Tor configuration and auth token exchange necessary to
deploy a highly resilient, network-agnostic private backplane to connect
infrastructure.

## What kind of services?

Text-based protocols (email, messaging), sensors & actuators, home automation,
and file synchronization are just some ideas to get you started.

In general, services that require little bandwidth or tolerate latency. With
ormesh, they can be accessed without the hassle of setting up iptables, NAT
port forwarding, VPNs, TLS, and without relying on central rendezvous servers.

# What ormesh isn't

ormesh is not a VPN in the conventional sense.

ormesh shares some properties with mesh overlay networks (NAT traversal,
end-to-end encryption, authenticated nodes), but it's not _mesh networking_ as
conventionally defined.

ormesh is not intended for operating unauthenticated anonymous hidden services.
Some measure of anonymity is an interesting side-effect of building on Tor, but
it is not a primary goal for ormesh. Future releases may provide the option to
reduce circuit length for improved network performance.

Low-latency, high bandwidth applications will probably not perform well over
Tor.

Also keep in mind that Tor only supports TCP.

# Configuring

## Exporting local services

A machine, VM or container running a web application might export access to
HTTP and SSH running locally.

```
$ ormesh export add 22
$ ormesh export add 80
$ ormesh status
service:
  export:
    - 127.0.0.1:22
    - 127.0.0.1:80
```

## Adding clients

Next, generate a token string that a client can import to access this server.
This string should be securely sent to the user of `my-MacBook` who is granted
access.

```
$ ormesh client add my-MacBook
Y5Cfw7A5RhP8Rd7xGYfD8N4oyEBpBWNR+6Qkgrbepk0=
```

Status shows the ports exported and the clients authorized to access them.

```
$ ormesh status
service:
  export:
    - 127.0.0.1:22
    - 127.0.0.1:80
  clients:
    - name: my-MacBook
      address: q6jo2z3bw5exkece.onion
```

# Consuming services

## Add a remote service, with client authentication

On the machine `my-MacBook`, add a remote using the authentication string displayed by
`client add` above.

```
$ ormesh remote add website Y5Cfw7A5RhP8Rd7xGYfD8N4oyEBpBWNR+6Qkgrbepk0=
```

```
$ ormesh status
remotes:
  - name: website
    address: q6jo2z3bw5exkece.onion
```

```
$ ormesh remote show website
q6jo2z3bw5exkece.onion
```

## Display an SSH config entry

Display an ssh-config(5) stanza for the remote.

```
$ ormesh remote ssh-config website
Host website
  ProxyCommand nc -X 5 -x localhost:9250 %h %p
  Hostname q6jo2z3bw5exkece.onion
```

## Importing remote services

Local port forwarding to the remote service. Local port 10022 will forward to
port 22 on the remote.

```
$ ormesh import add website 22 127.0.0.1:10022
$ ormesh status
remotes:
  - name: website
    address: q6jo2z3bw5exkece.onion
    imports:
    - local-addr: 127.0.0.1:10022
      remote-port: 22
```

Public ingress to a remote service. Useful for circumventing inbound port
blocks where the service is running. For example, you want to physically locate
your email server at home, but your ISP blocks SMTP and you lack a static IP
address. Import the service from a cloud instance with a public IP and DNS.

```
$ ormesh import add mailinabox 25 0.0.0.0:25
$ ormesh import add mailinabox 80 0.0.0.0:80
$ ormesh import add mailinabox 443 0.0.0.0:443
$ ormesh import add mailinabox 993 0.0.0.0:993
```

# Operating

An ormesh agent launches tor and operates it automatically, based on ormesh
configuration.

## Launching manually

```
$ ormesh agent run
```

will run the agent as configured above. Configuration changes made while the agent
is running are applied immediately.

## Setting up systemd

Display a systemd unit file that will run ormesh, from its current installed
binary path.

```
$ ormesh agent systemd-unit --user
[Unit]
Description=ormesh - onion-routed mesh

[Service]
ExecStart=/home/ubuntu/bin/ormesh agent run
Restart=always

[Install]
WantedBy=default.target
```
