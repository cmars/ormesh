// +build linux

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
