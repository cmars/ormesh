orc - Onion router control protocol library.
============================================

    go get github.com/sycamoreone/orc/control
    go get github.com/sycamoreone/orc/tor

Only some low-level functionality at the moment.

To do anything with this library you will have to read the
[control protocol specification](https://gitweb.torproject.org/torspec.git/tree/control-spec.txt)
beforehand.

Some documentation is [available](http://godoc.org/github.com/sycamoreone/orc).

Examples
---------

Some examples assume that a Tor router with ControlPort 9051 open and protected
by a password is running on localhost. You can start such a router temporarily
to run an example:

    > /usr/sbin/tor -f examples/torrc
    > go run examples/circuits/main.go

Other examples like `orc/examples/resolve/main.go` start there own Tor process
and only assume that a `tor` binary is available in $PATH.

Plans
----

I am planing to add to the library according to my own interests.
Still, if anybody is interested in a particular feature, please tell me!

Comments and suggestions are also appreciated!
