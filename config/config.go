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
	TorBinaryPath string
	SocksAddr     string
	ControlAddr   string
}

func ReadFile(fpath string) (*Config, error) {
	var config Config
	_, err := toml.DecodeFile(fpath, &config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config %q", fpath)
	}
	config.Path = fpath
	config.Dir = filepath.Dir(fpath)
	config.platformDefaults()
	return &config, nil
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
