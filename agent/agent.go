package agent

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
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
	torrcPath := filepath.Join(cfg.Dir, "tor", "torrc")
	if _, err := os.Stat(torrcPath); err != nil && os.IsNotExist(err) {
		err := ioutil.WriteFile(torrcPath, []byte(`
Log notice stdout
`), 0600)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create torrc")
		}
	}
	cmd := exec.Command(cfg.Node.Agent.TorBinaryPath,
		"-f", torrcPath,
		"--Log", "notice stdout",
		"--SocksPort", cfg.Node.Agent.SocksAddr,
		"--ControlPort", cfg.Node.Agent.ControlAddr,
		"--DataDirectory", dataDir,
		//"--HiddenServiceDir", hiddenServiceDir,
		"--HashedControlPassword", controlHash,
	)
	cmd.Dir = dataDir
	cmd.Stderr = os.Stderr
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
	return password, strings.TrimSpace(out.String()), nil
}

func (a *Agent) Start() error {
	log.Printf("%#v", a.cmd)
	stdout, err := a.cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to configure standard output")
	}
	err = a.cmd.Start()
	if err != nil {
		return errors.Wrap(err, "failed to start")
	}
	deadline := time.Now().Add(30 * time.Second)
	// Read output until one gets a "Bootstrapped 100%: Done" notice.
	buf := bufio.NewReader(stdout)
	line, err := buf.ReadString('\n')
	for err == nil {
		log.Println(line)
		if time.Now().After(deadline) {
			_ = a.cmd.Process.Kill()
			return errors.New("timeout waiting for tor to start")
		}
		if strings.Contains(line, "Bootstrapped 100%: Done") {
			break
		}
		line, err = buf.ReadString('\n')
	}
	if err != nil {
		return errors.Errorf("failed to read output: %v", err)
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
				fmt.Sprintf(`HiddenServiceDir="%s"`, a.hiddenServiceDir),
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
