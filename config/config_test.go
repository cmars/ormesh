package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func tempFile(t *testing.T) string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("tempfile: %v", err)
	}
	defer f.Close()
	return f.Name()
}

func TestConfigDefaults(t *testing.T) {
	fpath := tempFile(t)
	defer os.Remove(fpath)
	cfg, err := ReadFile(fpath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if cfg.Node.Agent.SocksPort != 9250 {
		t.Errorf("default node.agent.socksport = %d", cfg.Node.Agent.SocksPort)
	}
	if cfg.Node.Agent.ControlPort != 9251 {
		t.Errorf("default node.agent.controlport = %d", cfg.Node.Agent.ControlPort)
	}
}

func TestRoundTrip(t *testing.T) {
	fpath := tempFile(t)
	defer os.Remove(fpath)
	config := Config{
		Node: Node{
			Agent: Agent{
				TorBinary:   "/usr/bin/tor",
				TorDataDir:  "/var/lib/tor",
				SocksPort:   9050,
				ControlPort: 9051,
			},
			Service: Service{
				Exports: []Export{{
					Addr: "127.0.0.1",
					Port: 80,
				}},
				Clients: []Client{{
					Name:    "bob",
					Address: "qwertyuiop.onion",
				}},
			},
		},
	}
	err := WriteFile(&config, fpath)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	config2, err := ReadFile(fpath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	assert.Equal(t, &config, config2)
}
