// Copyright 2026 Google LLC
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

import "google.golang.org/adk/agent"

// Loader allows to load a particular agent by name and get the root agent
type Loader interface {
	// ListAgents returns a list of names of all agents
	ListAgents() []string
	// LoadAgent returns an agent by its name. Returns error if there is no agent with such a name.
	LoadAgent(name string) (agent.Agent, error)
	// RootAgent returns the root agent
	RootAgent() agent.Agent
}
