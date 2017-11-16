package tor

import (
	"time"
	"errors"
	"fmt"
	"strings"
	"strconv"
)

// A Config struct is used to configure a to be executed Tor process.
type Config struct {
	// Path is the path to a tor executable to be run. If path is the empty string,
	// $PATH is used to search for a tor executable.
	Path string
	// Timeout is the maximum amount of time we will wait for
	// a connect to the Tor network to complete.
	Timeout time.Duration
	// Options is a map of configuration options to values to be used
	// as command line arguments or in a torrc configuration file.
	Options map[string]string
	err     error
}

func NewConfig() *Config {
	c := &Config{
		Path: "",
		Options: make(map[string]string),
		err: nil,
	}
	return c
}

func (c *Config) setErr(format string, a ...interface{}) {
	err := errors.New(fmt.Sprintf(format, a...))
	if c.err == nil {
		c.err = err
	}
}

func (c *Config) Set(option string, value interface{}) {
	switch v := value.(type) {
	case int:
		c.Options[option] = strconv.Itoa(v)
	case string:
		c.Options[option] = dquote(v)
	default:
		c.setErr("value %v for option %s is not a string or int", value, option)
	}
}

// dquote returns s quoted in double-quotes, if it isn't already quoted and contains a space.
// Otherwise it just returns s itself.
func dquote(s string) string {
	if s[0] == '"' && s[len(s)-1] == '"' {
		// TODO check if there is a " in between the quotes that is not escaped using \
		return s
	}
	if strings.ContainsRune(s, ' ') {
		return "\"" + s + "\""
	}
	return s
}

// Err reports the first error that was encountered during the preceding calls to Set()
// and clears the saved error value to nil.
func (c *Config) Err() error {
	err := c.err
	c.err = nil
	return err
}

func (c Config) ToCmdLineFormat() []string {
	args := make([]string, 0)
	for k, v := range c.Options {
		args = append(args, "--"+k)
		args = append(args, v)
	}
	return args
}
