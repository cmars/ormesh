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

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

type Config struct {
	Node Node
	Dir  string
	Path string
}

type Node struct {
	Service Service
	Remotes []Remote
	Agent   Agent
}

type Service struct {
	Exports []string
	Clients []Client
}

type Client struct {
	Name    string
	Address string
}

type Remote struct {
	Name    string
	Address string
	Auth    string
	Imports []Import
}

type Import struct {
	LocalAddr  string
	LocalPort  int
	RemotePort int
}

type Agent struct {
	TorBinaryPath  string
	TorrcPath      string
	TorDataDir     string
	TorServicesDir string
	SocksAddr      string
	ControlAddr    string
	ControlCookie  string
	UseTorBrowser  bool
}

func (c *Config) defaults(md *toml.MetaData) {
	if c.Node.Agent.TorDataDir == "" && !md.IsDefined("Node", "Agent", "TorDataDir") {
		c.Node.Agent.TorDataDir = filepath.Join(c.Dir, "tor", "data")
	}
	if c.Node.Agent.TorrcPath == "" && !md.IsDefined("Node", "Agent", "TorrcPath") {
		c.Node.Agent.TorrcPath = filepath.Join(c.Node.Agent.TorDataDir, "torrc")
	}
	if c.Node.Agent.TorServicesDir == "" && !md.IsDefined("Node", "Agent", "TorServicesDir") {
		c.Node.Agent.TorServicesDir = filepath.Join(c.Node.Agent.TorDataDir, "services")
	}
	if c.Node.Agent.SocksAddr == "" && !md.IsDefined("Node", "Agent", "SocksAddr") {
		c.Node.Agent.SocksAddr = "127.0.0.1:9250"
	}
	if c.Node.Agent.ControlAddr == "" && !md.IsDefined("Node", "Agent", "ControlAddr") {
		c.Node.Agent.ControlAddr = "127.0.0.1:9251"
	}
}

func ReadFile(fpath string) (*Config, error) {
	var cfg Config
	md, err := toml.DecodeFile(fpath, &cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config %q", fpath)
	}
	cfg.init(fpath, &md)
	return &cfg, nil
}

func NewFile(fpath string) *Config {
	cfg := Config{}
	cfg.init(fpath, &toml.MetaData{})
	return &cfg
}

func (c *Config) init(fpath string, md *toml.MetaData) {
	c.Path = fpath
	c.Dir = filepath.Dir(fpath)
	c.platformDefaults()
	c.defaults(&toml.MetaData{})
}

func WriteFile(config *Config, fpath string) error {
	f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrapf(err, "failed to open %q for writing", fpath)
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	err = enc.Encode(config)
	if err != nil {
		return errors.Wrapf(err, "failed to encode config")
	}
	return nil
}
