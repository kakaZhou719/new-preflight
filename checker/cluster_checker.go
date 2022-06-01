// Copyright © 2022 Alibaba Group Holding Ltd.
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

package checker

import (
	"fmt"
	"github.com/pkg/errors"
	"preflight/checker/cluster/api/types"
	"preflight/checker/cluster/pkg/conf"
	"preflight/checker/cluster/pkg/run"
	"strings"
)

type ClusterCheck struct {
	AuthInfo types.ClusterInfoBrief
}

func (m ClusterCheck) Type() string {
	return strings.ToLower("ClusterCheck")
}

func (m ClusterCheck) PrettyName() string {
	return fmt.Sprintf("%s", m.Type())
}

func (ClusterCheck) Metadata() Metadata {
	return Metadata{
		Description: "Check the required os info and hardware resource of remote host",
		Level:       FatalLevel,
		Explain:     "if host hardware resource not meet required ,program may not run normally, or be very slow.",
		Suggestion:  "Maybe you should upgrade your machine",
	}
}

func (m ClusterCheck) Validate() (bool, error) {
	err := parseConfigs(m.AuthInfo)
	if err != nil {
		return false, err
	}
	ok, err := run.ValidateAll(&m.AuthInfo)
	if !ok {
		return false, errors.Errorf("failed to Validate ClusterInfo %v", err)
	}
	return true, nil
}

func parseConfigs(briefInfo types.ClusterInfoBrief) error {
	if briefInfo.SshUser == "" {
		briefInfo.SshUser = "root"
	}
	conf.SshConfig.User = briefInfo.SshUser

	conf.SshConfig.Password = briefInfo.SshPassword
	env := make(map[string]string)
	if len(briefInfo.Hosts) == 0 {
		return fmt.Errorf("hosts must be config in cluster-info-brief.yaml")
	} else {
		var hostsWithoutPort []string
		for _, host := range briefInfo.Hosts {
			hostsWithoutPort = append(hostsWithoutPort, run.IPFormat(host))
		}
		env["SSHHosts"] = fmt.Sprintf("(%s)", strings.Join(hostsWithoutPort, " "))
		env["SSHPort"] = "22"
	}

	conf.SshConfig.Env = env

	return nil
}
