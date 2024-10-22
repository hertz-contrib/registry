/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package polaris

import (
	polarishertz "github.com/cloudwego-contrib/cwgo-pkg/registry/polaris/polarishertz"
)

// Registry is extension interface of Hertz registry.Registry.
type Registry = polarishertz.Registry

// NewPolarisRegistry creates a polaris based registry.
func NewPolarisRegistry(configFile ...string) (Registry, error) {
	return polarishertz.NewPolarisRegistry(configFile...)
}
