// Copyright Istio Authors
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

package main

import (
	"dubbo-mesh/pilot/cmd/pilot-agent/app"
	"k8s.io/klog/v2"
	"os"
)

// TODO: get the config and bootstrap from istiod, by passing the env

// Use env variables - from injection, k8s and local namespace config map.
// No CLI parameters.
func main() {
	rootCmd := app.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		klog.Error(err)
		os.Exit(-1)
	}
}
