package agent

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cmars/orc/control"
	"github.com/pkg/errors"

	"github.com/cmars/ormesh/config"
)

type Agent struct {
	dataDir          string
	hiddenServiceDir string
	controlAddr      string
	controlPass      string
	conn             *control.Conn
	cmd              *exec.Cmd
}

func New(cfg *config.Config) (*Agent, error) {
	dataDir := filepath.Join(cfg.Dir, "tor", "data")
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %q", dataDir)
	}
	hiddenServiceDir := filepath.Join(cfg.Dir, "tor", "services")
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %q", hiddenServiceDir)
	}
	controlPass, controlHash, err := generateControlPass(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate control password hash")
	}
	cmd := exec.Command(cfg.Node.Agent.TorBinaryPath,
		"-f", filepath.Join(cfg.Dir, "tor", "torrc"),
		"--SocksPort", cfg.Node.Agent.SocksAddr,
		"--ControlPort", cfg.Node.Agent.ControlAddr,
		"--DataDirectory", dataDir,
		"--HiddenServiceDir", hiddenServiceDir,
		"--HashedControlPassword", controlHash,
	)
	cmd.Dir = dataDir
	return &Agent{
		dataDir:          dataDir,
		hiddenServiceDir: hiddenServiceDir,
		controlAddr:      cfg.Node.Agent.ControlAddr,
		controlPass:      controlPass,
		cmd:              cmd,
	}, nil
}

func generateControlPass(cfg *config.Config) (string, string, error) {
	var binpass [32]byte
	_, err := rand.Reader.Read(binpass[:])
	if err != nil {
		return "", "", errors.Wrap(err, "failed to generate password")
	}
	password := base64.URLEncoding.EncodeToString(binpass[:])
	cmd := exec.Command(cfg.Node.Agent.TorBinaryPath, "--hash-password", password)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to obtain hashed control password")
	}
	return password, out.String(), nil
}

func (a *Agent) Start() error {
	err := a.cmd.Start()
	if err != nil {
		return errors.Wrap(err, "failed to start")
	}
	conn, err := control.Dial(a.controlAddr)
	if err != nil {
		return errors.Wrap(err, "control connect failed")
	}
	err = conn.Auth(a.controlPass)
	if err != nil {
		return errors.Wrap(err, "control auth failed")
	}
	a.conn = conn
	return nil
}

func (a *Agent) Stop() error {
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
	for _, export := range svc.Exports {
		_, port, err := net.SplitHostPort(export)
		if err != nil {
			return errors.Wrapf(err, "invalid export %q", export)
		}
		_, err = a.conn.Send(control.Cmd{
			Keyword: "SETCONF",
			Arguments: []string{
				fmt.Sprintf(`HiddenServicePort="%s %s"`, port, export),
			},
		})
		if err != nil {
			return errors.Wrap(err, "failed to update HiddenServicePort")
		}
	}
	var clientNames []string
	for _, client := range svc.Clients {
		clientNames = append(clientNames, client.Name)
	}
	_, err := a.conn.Send(control.Cmd{
		Keyword: "SETCONF",
		Arguments: []string{
			fmt.Sprintf(`HiddenServiceAuthorizeClient="stealth %s"`, strings.Join(clientNames, ",")),
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to configure HiddenServiceAuthorizeClient")
	}
	_, err = a.conn.Send(control.Cmd{
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
