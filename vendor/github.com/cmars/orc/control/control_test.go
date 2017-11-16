package control

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

// Tests are somewhat inspired by Go's net/smtp's tests.

type faker struct {
	io.ReadWriter
}

func (f faker) Close() error                     { return nil }
func (f faker) LocalAddr() net.Addr              { return nil }
func (f faker) RemoteAddr() net.Addr             { return nil }
func (f faker) SetDeadline(time.Time) error      { return nil }
func (f faker) SetReadDeadline(time.Time) error  { return nil }
func (f faker) SetWriteDeadline(time.Time) error { return nil }

var basicCmds = `AUTHENTICATE "password"
XXX
MAPADDRESS 0.0.0.0=torproject.org 1.2.3.4=tor.freehaven.net
GETINFO version desc/name/moria1
`
var basicReplies = `250 OK
510 Unrecognized command "XXX"
250-127.192.10.10=torproject.org
250 1.2.3.4=tor.freehaven.net
250+desc/name/moria=
router moria1 128.31.0.34 9101 0 9131
..  // dot has to be escaped by another dot
fingerprint 9695 DFC3 5FFE B861 329B 9F1A B04C 4639 7020 CE31
uptime 1130120
// A dot on its own signals the end of the CmdData.
.
250-version=Tor 0.1.1.0-alpha-cvs
250 OK
`

// equal checks if two Reply structs are equal.
func equal(r, s Reply) bool {
	if r.Status != s.Status {
		return false
	}
	if r.Text != s.Text {
		return false
	}
	if len(r.Lines) != len(s.Lines) {
		return false
	}
	for i := range r.Lines {
		if r.Lines[i] != s.Lines[i] {
			return false
		}
	}
	return true
}

func TestSendReceive(t *testing.T) {
	client := strings.Join(strings.Split(basicCmds, "\n"), "\r\n")
	server := strings.Join(strings.Split(basicReplies, "\n"), "\r\n")

	var cmdbuf bytes.Buffer
	bcmdbuf := bufio.NewWriter(&cmdbuf)
	var fake faker
	fake.ReadWriter = bufio.NewReadWriter(bufio.NewReader(strings.NewReader(server)), bcmdbuf)
	c := Client(fake)

	if err := c.Auth("password"); err != nil {
		t.Fatalf("AUTHENTICATE failed: %s", err)
	}

	cmd := Cmd{Keyword: "XXX"}
	reply, err := c.Send(cmd)
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}
	expected := Reply{Status: 510, Text: "Unrecognized command \"XXX\""}
	if !equal(expected, *reply) {
		t.Fatalf("Got:\n%v\nExpected:\n%v", reply, expected)
	}

	cmd = Cmd{
		Keyword:   "MAPADDRESS",
		Arguments: []string{"0.0.0.0=torproject.org", "1.2.3.4=tor.freehaven.net"},
	}
	reply, err = c.Send(cmd)
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}
	expected = Reply{
		Status: 250,
		Text:   "1.2.3.4=tor.freehaven.net",
		Lines:  []ReplyLine{{Status: 250, Text: "127.192.10.10=torproject.org"}},
	}
	if !equal(expected, *reply) {
		t.Fatalf("Got:\n%v\nExpected:\n%v", reply, expected)
	}

	cmd = Cmd{
		Keyword:   "GETINFO",
		Arguments: []string{"version desc/name/moria1"},
	}
	reply, err = c.Send(cmd)
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}
	expected = Reply{
		Status: 250,
		Text:   "OK",
		Lines: []ReplyLine{
			{
				Status: 250,
				Text:   "desc/name/moria=",
				Data: `router moria1 128.31.0.34 9101 0 9131
.  // dot has to be escaped by another dot
fingerprint 9695 DFC3 5FFE B861 329B 9F1A B04C 4639 7020 CE31
uptime 1130120
// A dot on its own signals the end of the CmdData.
`,
			},
			{
				Status: 250,
				Text:   "version=Tor 0.1.1.0-alpha-cvs",
			},
		},
	}

	if !equal(expected, *reply) {
		t.Fatalf("Got:\n%v\nExpected:\n%v", reply, expected)
	}

	bcmdbuf.Flush()
	actualcmds := cmdbuf.String()
	if client != actualcmds {
		t.Fatalf("Got:\n%s\nExpected:\n%s", actualcmds, client)
	}
}
