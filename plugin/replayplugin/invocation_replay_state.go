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

package replayplugin

import "google.golang.org/adk/plugin/replayplugin/recording"

// InvocationReplayState tracks per-invocation replay state to isolate concurrent runs.
type InvocationReplayState struct {
	testCasePath     string
	userMessageIndex int
	recordings       *recording.Recordings

	// Per-agent replay indices for parallel execution
	// key: agent_name -> current replay index for that agent
	agentReplayIndices map[string]int
}

// NewInvocationReplayState behaves as the constructor.
func NewInvocationReplayState(testCasePath string, userMessageIndex int, recs *recording.Recordings) *InvocationReplayState {
	return &InvocationReplayState{
		testCasePath:       testCasePath,
		userMessageIndex:   userMessageIndex,
		recordings:         recs,
		agentReplayIndices: make(map[string]int),
	}
}

// GetTestCasePath returns the test case path.
func (s *InvocationReplayState) GetTestCasePath() string {
	return s.testCasePath
}

// GetUserMessageIndex returns the user message index.
func (s *InvocationReplayState) GetUserMessageIndex() int {
	return s.userMessageIndex
}

// GetRecordings returns the recordings object.
func (s *InvocationReplayState) GetRecordings() *recording.Recordings {
	return s.recordings
}

// GetAgentReplayIndex returns the index for the agent.
// In Go, looking up a missing key returns the zero value (0),
// so getOrDefault is intrinsic to the language for integers.
func (s *InvocationReplayState) GetAgentReplayIndex(agentName string) int {
	return s.agentReplayIndices[agentName]
}

// SetAgentReplayIndex sets the replay index for a specific agent.
func (s *InvocationReplayState) SetAgentReplayIndex(agentName string, index int) {
	s.agentReplayIndices[agentName] = index
}

// IncrementAgentReplayIndex increments the replay index for a specific agent.
func (s *InvocationReplayState) IncrementAgentReplayIndex(agentName string) {
	s.agentReplayIndices[agentName]++
}
