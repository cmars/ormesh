// Copyright © 2017 Casey Marshall
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cmars/orc/control"
	"github.com/pkg/errors"
	"golang.org/x/net/proxy"

	"github.com/cmars/ormesh/config"
)

type Agent struct {
	dataDir          string
	hiddenServiceDir string
	controlAddr      string
	conn             *control.Conn
	cmd              *exec.Cmd
	forwarders       []*forwarder
}

type forwarder struct {
	remoteAddr string
	remotePort int
	localAddr  string
	localPort  int
	dialer     proxy.Dialer
	l          *net.TCPListener
}

func New(cfg *config.Config) (*Agent, error) {
	if cfg.Node.Agent.UseTorBrowser {
		return newTorBrowserAgent(cfg)
	} else {
		return newStandaloneAgent(cfg)
	}
}

func newStandaloneAgent(cfg *config.Config) (*Agent, error) {
	dataDir := cfg.Node.Agent.TorDataDir
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %q", dataDir)
	}
	hiddenServiceDir := cfg.Node.Agent.TorServicesDir
	if err := os.MkdirAll(hiddenServiceDir, 0700); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %q", hiddenServiceDir)
	}
	torrcPath := filepath.Join(dataDir, "torrc")
	if _, err := os.Stat(torrcPath); err != nil && os.IsNotExist(err) {
		err := ioutil.WriteFile(torrcPath, []byte(`
Log notice stdout
`), 0600)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create torrc")
		}
	}
	args := []string{
		"-f", torrcPath,
		"--Log", "notice stderr",
		"--SocksPort", cfg.Node.Agent.SocksAddr,
		"--ControlPort", cfg.Node.Agent.ControlAddr,
		"--CookieAuthentication", "1",
		"--DataDirectory", dataDir,
	}
	cmd := exec.Command(cfg.Node.Agent.TorBinaryPath, args...)
	cmd.Dir = dataDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	dialer, err := proxy.SOCKS5("tcp", cfg.Node.Agent.SocksAddr, nil, proxy.Direct)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var forwarders []*forwarder
	for _, remote := range cfg.Node.Remotes {
		for _, import_ := range remote.Imports {
			forwarders = append(forwarders, &forwarder{
				dialer:     dialer,
				remoteAddr: remote.Address,
				remotePort: import_.RemotePort,
				localAddr:  import_.LocalAddr,
				localPort:  import_.LocalPort,
			})
		}
	}
	return &Agent{
		dataDir:          dataDir,
		hiddenServiceDir: hiddenServiceDir,
		controlAddr:      cfg.Node.Agent.ControlAddr,
		cmd:              cmd,
		forwarders:       forwarders,
	}, nil
}

func newTorBrowserAgent(cfg *config.Config) (*Agent, error) {
	hiddenServiceDir := cfg.Node.Agent.TorServicesDir
	if err := os.MkdirAll(hiddenServiceDir, 0700); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %q", hiddenServiceDir)
	}
	var forwarders []*forwarder
	dialer, err := proxy.SOCKS5("tcp", cfg.Node.Agent.SocksAddr, nil, proxy.Direct)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, remote := range cfg.Node.Remotes {
		for _, import_ := range remote.Imports {
			forwarders = append(forwarders, &forwarder{
				dialer:     dialer,
				remoteAddr: remote.Address,
				remotePort: import_.RemotePort,
				localAddr:  import_.LocalAddr,
				localPort:  import_.LocalPort,
			})
		}
	}
	return &Agent{
		dataDir:          cfg.Node.Agent.TorDataDir,
		hiddenServiceDir: hiddenServiceDir,
		controlAddr:      cfg.Node.Agent.ControlAddr,
		forwarders:       forwarders,
	}, nil
}

func (a *Agent) Start() error {
	var (
		conn *control.Conn
		err  error
	)
	if a.cmd != nil {
		err = a.cmd.Start()
		if err != nil {
			return errors.Wrap(err, "failed to start")
		}
	}
	// Try to connect for 45 seconds cumulative
	for s := 1; s < 10; s++ {
		controlCookie, err := ioutil.ReadFile(
			filepath.Join(a.dataDir, "control_auth_cookie"))
		if err != nil {
			time.Sleep(time.Duration(s) * time.Second)
			continue
		}
		conn, err = control.Dial(a.controlAddr)
		if err != nil {
			time.Sleep(time.Duration(s) * time.Second)
			continue
		}
		err = conn.AuthCookie(controlCookie)
		if err != nil {
			return errors.Wrap(err, "control auth failed")
		}
		a.conn = conn
		err = a.startForwarding()
		if err != nil {
			return errors.Wrap(err, "local imports failed to start")
		}
		return nil
	}
	return errors.Wrap(err, "control connect failed")
}

func (a *Agent) startForwarding() error {
	for i := range a.forwarders {
		err := a.forwarders[i].start()
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (f *forwarder) start() error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", f.localAddr, f.localPort))
	if err != nil {
		return errors.WithStack(err)
	}
	f.l = l.(*net.TCPListener)
	go f.accept()
	log.Printf("started listener %v", f.l.Addr())
	return nil
}

func (f *forwarder) accept() {
	for {
		c, err := f.l.Accept()
		if err != nil {
			log.Printf("listener exiting on error: %v", err)
			return
		}
		go f.handleConn(c.(*net.TCPConn))
	}
}

func (f *forwarder) handleConn(source *net.TCPConn) {
	log.Printf("connection from %s", source.RemoteAddr())
	source.SetKeepAlive(true)
	source.SetKeepAlivePeriod(time.Second * 60)
	log.Printf("dialing %s:%d", f.remoteAddr, f.remotePort)
	dest, err := f.dialer.Dial("tcp", fmt.Sprintf("%s:%d", f.remoteAddr, f.remotePort))
	if err != nil {
		log.Println(err)
		return
	}
	destTCP := dest.(*net.TCPConn)
	destTCP.SetKeepAlive(true)
	destTCP.SetKeepAlivePeriod(time.Second * 60)
	go f.forward(source, destTCP)
	f.forward(destTCP, source)
}

func (f *forwarder) forward(dest, source *net.TCPConn) {
	defer dest.CloseWrite()
	defer source.CloseRead()
	n, err := io.Copy(dest, source)
	if err != nil {
		log.Println(err)
	}
	log.Printf("copied %d bytes %v -> %v", n, source.RemoteAddr(), dest.RemoteAddr())
}

func (a *Agent) Stop() error {
	if a.cmd == nil {
		return nil
	}
	err := a.cmd.Process.Kill()
	if err != nil {
		return errors.Wrap(err, "failed to kill process")
	}
	err = a.cmd.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to wait for process exit")
	}
	return nil
}

func (a *Agent) UpdateServices(svc *config.Service) error {
	var setArgs, resetArgs []string

	if len(svc.Exports) > 0 {
		setArgs = append(setArgs, fmt.Sprintf(`HiddenServiceDir="%s"`, a.hiddenServiceDir))
		for _, export := range svc.Exports {
			setArgs = append(setArgs,
				fmt.Sprintf(`HiddenServicePort="%d %s"`, export.Port, export.LocalAddr))
		}
	} else {
		resetArgs = append(resetArgs, "HiddenServicePort")
	}

	var clientNames []string
	for _, client := range svc.Clients {
		clientNames = append(clientNames, client.Name)
	}
	if len(clientNames) > 0 {
		setArgs = append(setArgs,
			fmt.Sprintf(`HiddenServiceAuthorizeClient="stealth %s"`,
				strings.Join(clientNames, ",")))
	} else {
		resetArgs = append(resetArgs, "HiddenServiceAuthorizeClient")
	}

	if len(setArgs) > 0 {
		_, err := a.conn.Send(control.Cmd{
			Keyword:   "SETCONF",
			Arguments: setArgs,
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	if len(resetArgs) > 0 {
		_, err := a.conn.Send(control.Cmd{
			Keyword:   "RESETCONF",
			Arguments: resetArgs,
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	_, err := a.conn.Send(control.Cmd{
		Keyword:   "SAVECONF",
		Arguments: []string{},
	})
	if err != nil {
		return errors.Wrap(err, "failed to save configuration")
	}
	return nil
}

func (a *Agent) ClientAccess(clientName string) (string, string, error) {
	hostnamePath := filepath.Join(a.hiddenServiceDir, "hostname")
	f, err := os.Open(hostnamePath)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to open %q", hostnamePath)
	}
	defer f.Close()
	lines := bufio.NewScanner(f)
	for lines.Scan() {
		line := lines.Text()
		if strings.HasSuffix(line, fmt.Sprintf("# client: %s", clientName)) {
			fields := strings.Split(line, " ")
			if len(fields) < 5 {
				continue
			}
			return fields[0], fields[1], nil
		}
	}
	return "", "", errors.New("not found")
}

func (a *Agent) UpdateRemotes(node *config.Node) error {
	var args []string
	for _, remote := range node.Remotes {
		if remote.Auth != "" {
			args = append(args, fmt.Sprintf(`HidServAuth="%s %s"`, remote.Address, remote.Auth))
		}
	}
	if len(args) > 0 {
		_, err := a.conn.Send(control.Cmd{
			Keyword:   "SETCONF",
			Arguments: args,
		})
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		_, err := a.conn.Send(control.Cmd{
			Keyword:   "RESETCONF",
			Arguments: []string{"HidServAuth"},
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	_, err := a.conn.Send(control.Cmd{
		Keyword:   "SAVECONF",
		Arguments: []string{},
	})
	if err != nil {
		return errors.Wrap(err, "failed to save configuration")
	}
	return nil
}
