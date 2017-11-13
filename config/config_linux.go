// +build linux

package config

func (c *Config) platformDefaults() {
	if c.Node.Agent.TorBinaryPath == "" {
		c.Node.Agent.TorBinaryPath = "/usr/bin/tor"
	}
}
