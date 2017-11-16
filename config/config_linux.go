// +build linux

package config

import (
	"os"
	"path/filepath"
)

func (c *Config) platformDefaults() {
	if c.Node.Agent.TorBinaryPath == "" {
		if snapDir := os.Getenv("SNAP"); snapDir != "" {
			c.Node.Agent.TorBinaryPath = filepath.Join(snapDir, "usr", "bin", "tor")
		} else {
			c.Node.Agent.TorBinaryPath = "/usr/bin/tor"
		}
	}
}
