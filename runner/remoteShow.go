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
	"fmt"

	"github.com/pkg/errors"

	"github.com/cmars/ormesh/config"
)

type RemoteShow struct {
	Base
}

func (r *RemoteShow) Run(args []string) error {
	err := r.WithConfig(func(cfg *config.Config) error {
		remoteName := args[0]
		if !IsValidRemoteName(remoteName) {
			return errors.Errorf("invalid remote %q", remoteName)
		}
		for _, remote := range cfg.Node.Remotes {
			if remote.Name == remoteName {
				fmt.Printf("%#v\n", remote)
				return nil
			}
		}
		return errors.Errorf("no such remote %q", remoteName)
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
