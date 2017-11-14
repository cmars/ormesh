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
		"--HashedControlPassword", controlHash,
	)
	cmd.Dir = dataDir
	cmd.Stdout = os.Stdout
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
	var (
		conn *control.Conn
		err  error
	)

	log.Printf("%#v", a.cmd)
	err = a.cmd.Start()
	if err != nil {
		return errors.Wrap(err, "failed to start")
	}
	// Try to connect for 45 seconds cumulative
	for s := 1; s < 10; s++ {
		conn, err = control.Dial(a.controlAddr)
		if err != nil {
			time.Sleep(time.Duration(s) * time.Second)
			continue
		}
		err = conn.Auth(a.controlPass)
		if err != nil {
			return errors.Wrap(err, "control auth failed")
		}
		a.conn = conn
		return nil
	}
	return errors.Wrap(err, "control connect failed")
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
	if len(svc.Exports) == 0 {
		return nil
	}
	args := []string{fmt.Sprintf(`HiddenServiceDir="%s"`, a.hiddenServiceDir)}
	for _, export := range svc.Exports {
		_, port, err := net.SplitHostPort(export)
		if err != nil {
			return errors.Wrapf(err, "invalid export %q", export)
		}
		args = append(args,
			fmt.Sprintf(`HiddenServicePort="%s %s"`, port, export))
	}
	var clientNames []string
	for _, client := range svc.Clients {
		clientNames = append(clientNames, client.Name)
	}
	if len(clientNames) > 0 {
		args = append(args,
			fmt.Sprintf(`HiddenServiceAuthorizeClient="stealth %s"`,
				strings.Join(clientNames, ",")))
	}
	log.Println(args)
	_, err := a.conn.Send(control.Cmd{
		Keyword:   "SETCONF",
		Arguments: args,
	})
	if err != nil {
		return errors.Wrap(err, "failed to configure hidden services")
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
		log.Println(line)
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
