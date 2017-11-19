package agent

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
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
	controlPass      string
	conn             *control.Conn
	cmd              *exec.Cmd
	importers        []importer
}

type importer struct {
	config.Import
	RemoteAddr string
	SocksAddr  string
	l          net.Listener
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
	geoIPFile := filepath.Join(cfg.Dir, "tor", "geoip")
	geoIPv6File := filepath.Join(cfg.Dir, "tor", "geoip6")
	cmd := exec.Command(cfg.Node.Agent.TorBinaryPath,
		"-f", torrcPath,
		"--Log", "notice stdout",
		"--SocksPort", cfg.Node.Agent.SocksAddr,
		"--ControlPort", cfg.Node.Agent.ControlAddr,
		"--HashedControlPassword", controlHash,
		"--DataDirectory", dataDir,
		"--GeoIPFile", geoIPFile,
		"--GeoIPv6File", geoIPv6File,
	)
	cmd.Dir = dataDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	var importers []importer
	for _, remote := range cfg.Node.Remotes {
		for _, import_ := range remote.Imports {
			importers = append(importers, importer{
				RemoteAddr: remote.Address,
				Import:     import_,
				SocksAddr:  cfg.Node.Agent.SocksAddr,
			})
		}
	}
	return &Agent{
		dataDir:          dataDir,
		hiddenServiceDir: hiddenServiceDir,
		controlAddr:      cfg.Node.Agent.ControlAddr,
		controlPass:      controlPass,
		cmd:              cmd,
		importers:        importers,
	}, nil
}

func generateControlPass(cfg *config.Config) (string, string, error) {
	var binpass [32]byte
	_, err := rand.Reader.Read(binpass[:])
	if err != nil {
		return "", "", errors.Wrap(err, "failed to generate password")
	}
	password := base64.URLEncoding.EncodeToString(binpass[:])
	// On windows, geoip files need to be specified with an absolute path or
	// they generate warnings that get mixed up with stdout.
	geoIPFile := filepath.Join(cfg.Dir, "tor", "geoip")
	geoIPv6File := filepath.Join(cfg.Dir, "tor", "geoip6")
	cmd := exec.Command(cfg.Node.Agent.TorBinaryPath,
		"--GeoIPFile", geoIPFile,
		"--GeoIPv6File", geoIPv6File,
		"--hash-password", password,
	)
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
		err = a.startImports()
		if err != nil {
			return errors.Wrap(err, "local imports failed to start")
		}
		return nil
	}
	return errors.Wrap(err, "control connect failed")
}

func (a *Agent) startImports() error {
	for _, i := range a.importers {
		err := i.start()
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (i *importer) start() error {
	var err error
	i.l, err = net.Listen("tcp", fmt.Sprintf("%s:%d", i.LocalAddr, i.LocalPort))
	if err != nil {
		return errors.WithStack(err)
	}
	dialer, err := proxy.SOCKS5("tcp", i.SocksAddr, nil, proxy.Direct)
	if err != nil {
		return errors.WithStack(err)
	}
	handle := func(conn net.Conn) {
		defer conn.Close()
		remoteConn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", i.RemoteAddr, i.RemotePort))
		if err != nil {
			log.Println(err)
			return
		}
		defer remoteConn.Close()
		fwd1 := make(chan struct{})
		go func() {
			_, err := io.Copy(conn, remoteConn)
			if err != nil {
				log.Println(err)
			}
			close(fwd1)
		}()
		fwd2 := make(chan struct{})
		go func() {
			_, err := io.Copy(remoteConn, conn)
			if err != nil {
				log.Println(err)
			}
			close(fwd2)
		}()
		select {
		case <-fwd1:
		case <-fwd2:
		}
	}
	go func() {
		for {
			conn, err := i.l.Accept()
			if err != nil {
				log.Println(err)
				return
			}
			go handle(conn)
		}
	}()
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
	var setArgs, resetArgs []string

	if len(svc.Exports) > 0 {
		setArgs = append(setArgs, fmt.Sprintf(`HiddenServiceDir="%s"`, a.hiddenServiceDir))
		for _, export := range svc.Exports {
			_, port, err := net.SplitHostPort(export)
			if err != nil {
				return errors.Wrapf(err, "invalid export %q", export)
			}
			setArgs = append(setArgs,
				fmt.Sprintf(`HiddenServicePort="%s %s"`, port, export))
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
