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

package runner

import (
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/cmars/ormesh/config"
)

type Runner interface {
	InitConfig() error
	Run(args []string) error
	WithConfig(f func(*config.Config) error) error
	WithConfigForUpdate(f func(*config.Config) error) error
}

func Run(r Runner, args []string) error {
	err := r.InitConfig()
	if err != nil {
		return errors.WithStack(err)
	}
	err = r.Run(args)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type Base struct {
	Runner
	ConfigFile string
}

func (r *Base) InitConfig() error {
	return InitConfig(&r.ConfigFile)
}

func InitConfig(configFile *string) error {
	if *configFile == "" {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			return errors.Wrapf(err, "failed to locate home directory")
		}
		ormeshDir := filepath.Join(home, ".ormesh")
		if err := os.MkdirAll(ormeshDir, 0700); err != nil {
			return errors.Wrapf(err, "failed to create %q", ormeshDir)
		}
		*configFile = filepath.Join(ormeshDir, "config")
	}
	return nil
}

func (r *Base) WithConfig(f func(*config.Config) error) error {
	cfg, err := config.ReadFile(r.ConfigFile)
	if os.IsNotExist(errors.Cause(err)) {
		cfg, err = config.NewFile(r.ConfigFile)
		if err != nil {
			return errors.WithStack(err)
		}
	} else if err != nil {
		return errors.WithStack(err)
	}
	err = f(cfg)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Base) WithConfigForUpdate(f func(*config.Config) error) error {
	cfg, err := config.ReadFile(r.ConfigFile)
	if os.IsNotExist(errors.Cause(err)) {
		cfg, err = config.NewFile(r.ConfigFile)
		if err != nil {
			return errors.WithStack(err)
		}
	} else if err != nil {
		return errors.WithStack(err)
	}
	err = f(cfg)
	if err != nil {
		return errors.WithStack(err)
	}
	err = config.WriteFile(cfg, r.ConfigFile)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
