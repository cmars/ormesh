// +build darwin

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
