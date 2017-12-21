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
	"strings"

	"github.com/pkg/errors"

	"github.com/cmars/ormesh/config"
)

type RemoteAdd struct {
	Base
}

func (r *RemoteAdd) Run(args []string) error {
	err := r.WithConfigForUpdate(func(cfg *config.Config) error {
		remoteName, remoteAddr, clientAuth := args[0], args[1], args[2]
		if !IsValidRemoteName(remoteName) {
			return errors.Errorf("invalid remote name %q", remoteName)
		}
		if !strings.HasSuffix(remoteAddr, ".onion") {
			return errors.Errorf("invalid remote addr %q", remoteAddr)
		}
		for i := range cfg.Node.Remotes {
			if cfg.Node.Remotes[i].Name == remoteName {
				return errors.Errorf("remote %q already exists", remoteName)
			}
		}
		remote := config.Remote{
			Name:    remoteName,
			Address: remoteAddr,
			Auth:    clientAuth,
		}
		cfg.Node.Remotes = append(cfg.Node.Remotes, remote)
		return nil
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
