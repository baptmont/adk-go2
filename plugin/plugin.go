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

// Package plugin provides.
package plugin

import (
	"google.golang.org/genai"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/session"
)

type Config struct {
	Name string

	OnUserMessageCallback OnUserMessageCallback

	OnEventCallback OnEventCallback

	BeforeRunCallback BeforeRunCallback
	AfterRunCallback  AfterRunCallback

	BeforeAgentCallback agent.BeforeAgentCallback
	AfterAgentCallback  agent.AfterAgentCallback

	BeforeModelCallback  llmagent.BeforeModelCallback
	AfterModelCallback   llmagent.AfterModelCallback
	OnModelErrorCallback llmagent.OnModelErrorCallback

	BeforeToolCallback  llmagent.BeforeToolCallback
	AfterToolCallback   llmagent.AfterToolCallback
	OnToolErrorCallback llmagent.OnToolErrorCallback

	CloseFunc func() error
}

func New(cfg Config) (*Plugin, error) {
	p := &Plugin{
		Name:                  cfg.Name,
		OnUserMessageCallback: cfg.OnUserMessageCallback,
		OnEventCallback:       cfg.OnEventCallback,
		BeforeRunCallback:     cfg.BeforeRunCallback,
		AfterRunCallback:      cfg.AfterRunCallback,
		BeforeAgentCallback:   cfg.BeforeAgentCallback,
		AfterAgentCallback:    cfg.AfterAgentCallback,
		BeforeModelCallback:   cfg.BeforeModelCallback,
		AfterModelCallback:    cfg.AfterModelCallback,
		OnModelErrorCallback:  cfg.OnModelErrorCallback,
		BeforeToolCallback:    cfg.BeforeToolCallback,
		AfterToolCallback:     cfg.AfterToolCallback,
		OnToolErrorCallback:   cfg.OnToolErrorCallback,
		// Map the exported config field to the unexported plugin field
		closeFunc: cfg.CloseFunc,
	}

	// Ensure closeFunc is never nil so p.Close() doesn't panic
	if p.closeFunc == nil {
		p.closeFunc = func() error {
			return nil
		}
	}

	return p, nil
}

type Plugin struct {
	Name                  string
	OnUserMessageCallback OnUserMessageCallback

	OnEventCallback OnEventCallback

	BeforeRunCallback BeforeRunCallback
	AfterRunCallback  AfterRunCallback

	BeforeAgentCallback agent.BeforeAgentCallback
	AfterAgentCallback  agent.AfterAgentCallback

	BeforeModelCallback  llmagent.BeforeModelCallback
	AfterModelCallback   llmagent.AfterModelCallback
	OnModelErrorCallback llmagent.OnModelErrorCallback

	BeforeToolCallback  llmagent.BeforeToolCallback
	AfterToolCallback   llmagent.AfterToolCallback
	OnToolErrorCallback llmagent.OnToolErrorCallback

	closeFunc func() error
}

func (p *Plugin) Close() error {
	return p.closeFunc()
}

type OnUserMessageCallback func(agent.InvocationContext, *genai.Content) (*genai.Content, error)

type BeforeRunCallback func(agent.InvocationContext) (*genai.Content, error)

type AfterRunCallback func(agent.InvocationContext)

type OnEventCallback func(agent.InvocationContext, *session.Event) (*session.Event, error)
