# ormesh - onion-routed mesh

ormesh helps you build a private mesh of hosts connected through Tor.

Access your services running almost anywhere, from almost anywhere else.
Without much hassle.

Abstract away geography and network topologies.

Built on the security and durability of Tor and the Tor network.

## Secure

Services may only be accessed from authenticated clients. Unauthenticated users
cannot access, or even detect the existence of private services. An attacker
would need to break Tor or obtain your secrets in order to access your
services.

## Resilient

Being resistant to censorship also helps Tor traverse all kinds of networks
between your nodes. Interconnect them with no SPoFs.

# What ormesh is

ormesh forwards ports like SSH, but eliminates the need to set up networking
between the hosts. As long as the hosts can connect to Tor, they can connect to
each other.

ormesh is based on private hidden services with client authentication. This can
be done with lots of configuration; all ormesh does is help automate the work
of distributing authentication tokens and hidden service addresses.

Simple devops principles, to make interconnecting services across networks
easy, safe and fun.

## What kind of services?

Services for individuals and small groups often require little bandwidth, and
can be easily hosted in dense containers. With ormesh, they can be accessed
without the hassle of setting up iptables, NAT, VPNs, TLS.

# What ormesh isn't

ormesh is not a VPN in the conventional sense.

ormesh shares some properties with mesh overlay networks (NAT traversal,
end-to-end encryption, authenticated nodes), but it's not _mesh networking_ as
conventionally defined.

ormesh is not intended for operating unauthenticated anonymous hidden
services, nor is it intended for low-latency, high bandwidth applications.

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

Generate a token string that a client can use to import to access this server.
This string should be securely sent to the user of `my-MacBook` who is granted
access.

_TODO: For now, this string contains the encoded client secret. Eventually, this
should be replaced with a nonce used to obtain that secret._

```
$ ormesh client add my-MacBook
Y5Cfw7A5RhP8Rd7xGYfD8N4oyEBpBWNR+6Qkgrbepk0=
```

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
$ ormesh import website add 10022:127.0.0.1:22
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
$ ormesh import mailinabox add 25:0.0.0.0:25
$ ormesh import mailinabox add 80:0.0.0.0:80
$ ormesh import mailinabox add 443:0.0.0.0:443
$ ormesh import mailinabox add 993:0.0.0.0:993
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

## Checking agent status

```
$ ormesh agent status
agent:
  - state: down
```

## Setting up systemd

Display a systemd unit file that will run ormesh, from its current installed
binary path.

```
$ ormesh agent systemd-unit --user
[Unit]
Description=ormesh - onion-routed mesh

[Service]
ExecStart=/home/ubuntu/bin/ormesh agent run
Restart=on-failure
SuccessExitStatus=0
RestartForceExitStatus=0

[Install]
WantedBy=default.target
```
