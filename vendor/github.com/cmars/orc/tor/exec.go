// Package tor supplies helper functions to start a tor binary as a slave process.
package tor

import (
	"time"
	"os/exec"
	"io"
	"bufio"
	"strings"
	"errors"
)

// Cmd represents an tor executable to be run as a slave process.
type Cmd struct {
	Config *Config
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

// NewCmd returns a Cmd to run a tor process using the configuration values in config.
func NewCmd(config *Config) (*Cmd, error) {
	if config.Path == "" {
		file, err := exec.LookPath("tor")
		if err != nil {
			return nil, err
		}
		config.Path = file
	}

	cmd := exec.Command(config.Path, config.ToCmdLineFormat()...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	return &Cmd{
		Config: config,
		cmd: cmd,
		stdout: stdout,
	}, nil
}

func (c *Cmd) Start() error {
	deadline := time.Now().Add(c.Config.Timeout)
	err := c.cmd.Start()
	if err != nil {
		return err
	}

	// Read output until one gets a "Bootstrapped 100%: Done" notice.
	buf := bufio.NewReader(c.stdout)
	line, err := buf.ReadString('\n')
	for err == nil {
		if time.Now().After(deadline) {
			_ = c.cmd.Process.Kill()
			return errors.New("orc/tor: process killed because of timeout")
		}
		if strings.Contains(line, "Bootstrapped 100%: Done") {
			break
		}
		line, err = buf.ReadString('\n')
	}
	return nil
}

func (c *Cmd) Wait() error {
	return c.cmd.Wait()
}

func (c *Cmd) KillUnlessExited() {
	if c.cmd.ProcessState.Exited() {
		return
	}
	c.cmd.Process.Kill()
}