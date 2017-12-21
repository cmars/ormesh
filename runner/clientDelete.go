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
	"github.com/pkg/errors"

	"github.com/cmars/ormesh/config"
)

type ClientDelete struct {
	Base
}

func (r *ClientDelete) Run(args []string) error {
	err := r.WithConfigForUpdate(func(cfg *config.Config) error {
		clientName := args[0]
		if !IsValidClientName(clientName) {
			return errors.Errorf("invalid client name %q", clientName)
		}
		index := -1
		for i := range cfg.Node.Service.Clients {
			if cfg.Node.Service.Clients[i].Name == clientName {
				index = i
				break
			}
		}
		if index == -1 {
			return errors.Errorf("no such client %q", clientName)
		}
		cfg.Node.Service.Clients = append(
			cfg.Node.Service.Clients[:index],
			cfg.Node.Service.Clients[index+1:]...)
		return nil
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
