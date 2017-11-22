// +build darwin

package config

import (
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

func (c *Config) platformDefaults() {
	if c.Node.Agent.TorBinaryPath == "" {
		c.Node.Agent.TorBinaryPath = "/Applications/TorBrowser.app/Contents/MacOS/Tor/tor.real"
	}
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	if c.Node.Agent.TorDataDir == "" {
		c.Node.Agent.TorDataDir = filepath.Join(home,
			"Library", "Application Support", "TorBrowser-Data", "Tor")
		c.Node.Agent.UseTorBrowser = true
	}
	if c.Node.Agent.TorrcPath == "" {
		c.Node.Agent.TorrcPath = filepath.Join(c.Node.Agent.TorDataDir, "torrc")
	}
	if c.Node.Agent.TorServicesDir == "" {
		c.Node.Agent.TorServicesDir = filepath.Join(c.Node.Agent.TorDataDir, "services")
	}
	if c.Node.Agent.ControlCookie == "" {
		c.Node.Agent.ControlCookie = filepath.Join(c.Node.Agent.TorDataDir, "control_auth_cookie")
	}
	if c.Node.Agent.ControlAddr == "" {
		c.Node.Agent.ControlAddr = "127.0.0.1:9151"
	}
	if c.Node.Agent.SocksAddr == "" {
		c.Node.Agent.SocksAddr = "127.0.0.1:9150"
	}
}
