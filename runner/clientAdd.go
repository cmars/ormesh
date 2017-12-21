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
	"encoding/json"
	"fmt"
	"os"

	"github.com/mdp/qrterminal"
	"github.com/pkg/errors"

	"github.com/cmars/ormesh/agent"
	"github.com/cmars/ormesh/config"
)

type ClientAdd struct {
	Base
	DisplayQR bool
}

func (r *ClientAdd) Run(args []string) error {
	err := r.WithConfigForUpdate(func(cfg *config.Config) error {
		client, err := r.addClient(cfg, args[0])
		if err != nil {
			return errors.WithStack(err)
		}
		if r.DisplayQR {
			qrDoc := struct {
				AuthCookieValue string `json:"auth_cookie_value"`
				Domain          string `json:"domain"`
			}{
				AuthCookieValue: client.Auth,
				Domain:          client.Address,
			}
			qrText, err := json.Marshal(&qrDoc)
			if err != nil {
				return errors.WithStack(err)
			}
			qrterminal.Generate(string(qrText), qrterminal.H, os.Stdout)
		} else {
			fmt.Printf("%s %s\n", client.Address, client.Auth)
		}
		return nil
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *ClientAdd) addClient(cfg *config.Config, clientName string) (*config.Client, error) {
	if !IsValidClientName(clientName) {
		return nil, errors.Errorf("invalid client name %q", clientName)
	}
	index := -1
	for i := range cfg.Node.Service.Clients {
		if cfg.Node.Service.Clients[i].Name == clientName {
			index = i
			break
		}
	}
	if index < 0 {
		index = len(cfg.Node.Service.Clients)
		cfg.Node.Service.Clients = append(cfg.Node.Service.Clients, config.Client{
			Name: clientName,
		})
	}
	a, err := agent.New(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize agent")
	}
	err = a.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start agent")
	}
	defer a.Stop()
	err = a.UpdateServices(&cfg.Node.Service)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update tor hidden services")
	}
	address, clientAuth, err := a.ClientAccess(clientName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read tor client auth")
	}
	cfg.Node.Service.Clients[index].Address = address
	cfg.Node.Service.Clients[index].Auth = clientAuth
	return &cfg.Node.Service.Clients[index], nil
}
