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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"google.golang.org/genai"

	"google.golang.org/adk/session"
)

// UserContent wraps the base Content type with a 'user' role.
type UserContent struct {
	genai.Content
}

// ModelContent wraps the base Content type with a 'model' role.
type ModelContent struct {
	genai.Content
}

// UserMessage represents the individual messages in a test sequence.
type UserMessage struct {
	// Pydantic's Optional[str] is best represented as *string in Go
	// to distinguish between an empty string and a null/missing field.
	Text *string `yaml:"text,omitempty"`

	Content *UserContent `yaml:"content,omitempty"`

	// Optional[dict[str, Any]] maps to map[string]any
	StateDelta map[string]any `yaml:"state_delta,omitempty"`
}

// TestSpec defines the human-authored specification for conformance.
type TestSpec struct {
	Description string `yaml:"description"`
	Agent       string `yaml:"agent"`

	// Default factories in Python are handled by Go's natural
	// initialization of maps and slices to nil/empty.
	InitialState map[string]any `yaml:"initial_state"`
	UserMessages []UserMessage  `yaml:"user_messages"`
}

// TestCase represents a single conformance test case.
// In Go, we use a struct instead of a @dataclass.
type TestCase struct {
	Category string
	Name     string
	Dir      string // Using string to represent the path
	Spec     TestSpec
}

// LoadTestCase loads TestSpec from spec.yaml file.
func LoadTestCase(testCaseDir string) (*TestSpec, error) {
	specFile := filepath.Join(testCaseDir, "spec.yaml")

	data, err := os.ReadFile(specFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec TestSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse spec yaml: %w", err)
	}

	return &spec, nil
}

// LoadRecordedSession loads recorded session data from generated-session.yaml file.
func LoadRecordedSession(testCaseDir string) (*session.Session, error) {
	sessionFile := filepath.Join(testCaseDir, "generated-session.yaml")

	// Check if file exists (equivalent to session_file.exists())
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	// Handle empty file
	if len(data) == 0 {
		return nil, nil
	}

	var sess session.Session
	if err := yaml.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("failed to parse session yaml: %w", err)
	}

	return &sess, nil
}
