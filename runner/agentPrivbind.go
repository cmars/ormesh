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

package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

type AgentPrivbind struct {
	Base
}

func (r *AgentPrivbind) Run(args []string) error {
	binaryPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return errors.WithStack(err)
	}
	if os.Getuid() == 0 {
		cmd := exec.Command("setcap", "cap_net_bind_service=+ep", binaryPath)
		err := cmd.Run()
		if err != nil {
			return errors.Wrapf(err, "setcap failed")
		}
	} else {
		cmd := exec.Command("/bin/sh", "-c",
			fmt.Sprintf("sudo setcap 'cap_net_bind_service=+ep' %s", binaryPath))
		err := cmd.Run()
		if err != nil {
			return errors.Wrapf(err, "setcap failed")
		}
	}
	return nil
}
