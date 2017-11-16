package control

import (
	"errors"
	"strconv"
	"strings"
)

// Reply represents a reply send from the server to the client.
type Reply struct {
	Status int         // the StatusCode of the reply
	Text   string      // ReplyText of the EndReplyLine
	Lines  []ReplyLine // MidReplyLines and DataReplyLines
}

func (r Reply) String() string {
	s := strconv.Itoa(r.Status) + " " + r.Text
	for _, line := range r.Lines {
		s = s + "\n" + line.String()
	}
	return s
}

// Status codes of replies from the onion router.
const (
	StatusOK                   = 250
	StatusOperationUnnecessary = 251

	StatusResourceExhausted = 451

	StatusProtocolSyntaxError = 500

	StatusUnrecognizedCommand    = 510
	StatusUnimplementedCommand   = 511
	StatusArgumentSyntaxError    = 512
	StatusUnrecognizedArgument   = 513
	StatusAuthenticationRequired = 514
	StatusBadAuthentication      = 515

	StatusUnspecifiedError   = 550
	StatusInternalError      = 551
	StatusUnrecognizedEntity = 552

	StatusInvalidConfigurationValue = 553
	StatusInvalidDescriptor         = 554

	StatusUnmanagedEntity = 555

	StatusAsyncEventNotification = 650
)

// IsAsync returns true if r is an asynchronous reply and false otherwise.
func (r Reply) IsAsync() bool {
	return r.Status == 650
}

// IsSync returns true if r is an synchronous reply and false otherwise.
func (r Reply) IsSync() bool {
	return r.Status != 650
}

// ReplyLine represents a MidReplyLine or DataReplyLine read from the server.
type ReplyLine struct {
	Status int
	Text   string
	// Data is the empty string for MidReplyLines and the CmdData with
	// dot encoding removed for DataReplyLines.
	Data string
}

func (rl ReplyLine) String() string {
	s := strconv.Itoa(rl.Status) + " " + rl.Text + " " + rl.Data
	return s
}

// the possible kinds of ReplyLines
const (
	midLine = iota
	dataLine
	endLine
)

// readLine reads a MidLine, EndLine or DataLine into a ReplyLine struct.
// If r is nil, creates a new ReplyLine
func (c Conn) readLine() (lineType int, rl *ReplyLine, err error) {
	line, err := c.text.ReadLine()
	if err != nil {
		return
	}

	status, lineType, text, err := parseLine(line)
	if err != nil {
		return lineType, nil, err
	}
	rl = new(ReplyLine)
	rl.Status = status
	rl.Text = text

	if lineType == dataLine {
		data, err := c.readData()
		rl.Data = data
		if err != nil {
			return lineType, nil, err
		}
	}
	return
}

func parseLine(line string) (status, lineType int, text string, err error) {
	if len(line) < 4 || line[3] != ' ' && line[3] != '-' && line[3] != '+' {
		err = errors.New("protocol error: : " + line)
		return
	}
	switch line[3] {
	case '-':
		lineType = midLine
	case ' ':
		lineType = endLine
	case '+':
		lineType = dataLine

	}
	status, err = strconv.Atoi(line[0:3])
	if err != nil || status < 100 {
		err = errors.New("protocol errors: invalid status code: " + line)
		return
	}
	text = line[4:]
	return
}

// readData reads the dot encoded CmdData following a DataReplyLine.
// It returns a string with dot encoding removed.
func (c Conn) readData() (string, error) {
	buf, err := c.text.ReadDotBytes()
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// Receive reads and returns a single reply from the Tor server.
// It makes no distinction between synchronous and asynchronous replies.
func (c Conn) Receive() (*Reply, error) {
	// We read a multi-line reply containing
	// lines of the form
	//
	//	status-message line 1	// a MidReplyLine
	//
	//	status+message line 2	// a DataReplyLine
	//	<dot enstatusd data>
	//	.
	//
	//	status message line n	// a EndReplyLine
	//
	// into r. status is a three-digit status code The reply is terminated
	// by a EndReplyLine.
	reply := new(Reply)
	lineType, replyLine, err := c.readLine()
	if err != nil {
		return reply, nil
	}

	if lineType != endLine {
		if reply.Lines == nil {
			reply.Lines = make([]ReplyLine, 0, 1)
		} else {
			reply.Lines = reply.Lines[:0]
		}
	}
	for err == nil && lineType != endLine {
		// TODO: Should we check that the second Status isn't different from the first?
		reply.Lines = append(reply.Lines, *replyLine)
		lineType, replyLine, err = c.readLine()
		if err != nil {
			return reply, err
		}
	}

	// replyLine now contains the EndReplyLine
	reply.Status = replyLine.Status
	reply.Text = replyLine.Text
	return reply, nil
}

// ReceiveSync reads replies from the Tor server. It returns the first synchronous reply;
// replies read before that are send to the connections Replies channel.
// ReceiveSync blocks until the replies are read from that channel.
func (c Conn) ReceiveSync() (*Reply, error) {
	r, err := c.Receive()
	if err != nil {
		return r, err
	}
	for r.IsAsync() {
		c.Replies <- r
		r, err = c.Receive()
		if err != nil {
			return r, err
		}
	}
	return r, nil
}

// ReceiveToChan reads a single reply from the Tor server and sends
// it to the connections Replies channel. ReceiveToChan blocks until the
// reply is read from the channel.
func (c Conn) ReceiveToChan() error {
	r, err := c.Receive()
	if err != nil {
		return err
	}
	c.Replies <- r
	return nil
}

// A Handler can be registered to handle asynchronous replies send
// by an Tor router. It is customary for Handler
// to communicate back using channels.
type Handler func(r *Reply)

// Demux is a simple demultiplexer for asynchronous replies.
// It matches the type of a reply against a list of registered events
// and calls the corresponding Handler function.
type Demux struct {
	conn    *Conn
	handler map[string]Handler
}

// NewMux allocates and returns a new Demux.
func NewDemux(c *Conn) *Demux {
	return &Demux{conn: c, handler: make(map[string]Handler)}
}

// Handle registers f to handle replies of type event.
// The different events are listed in section 4.1 of the control-spec.
func (m *Demux) Handle(event string, f Handler) Handler {
	if event == "" {
		panic("control: invalid event " + event)
	}
	if f == nil {
		panic("control: nil Handler")
	}
	old, ok := m.handler[event]
	m.handler[event] = f
	if ok {
		return old
	}
	return nil
}

// Serve reads replies from the Replies channel of m's Conn and launches the
// corresponding Handler functions in new goroutines.
func (m *Demux) Serve() {
	for {
		r := <-m.conn.Replies
		event := r.Text[:strings.IndexByte(r.Text, ' ')]
		f, ok := m.handler[event]
		if !ok {
			continue
			// TODO: Would be nice to do some error reporting here; maybe using some errChan channel.
		}
		go f(r)
	}
}
