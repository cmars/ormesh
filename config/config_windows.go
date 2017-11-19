// +build windows

package config

import (
	"log"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

func (c *Config) platformDefaults() {
	if c.Node.Agent.TorBinaryPath == "" {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalf("failed to locate home directory: %v", err)
		}
		c.Node.Agent.TorBinaryPath = filepath.Join(home, "Desktop",
			"Tor Browser", "Browser", "TorBrowser", "Tor", "tor.exe")
	}
}
