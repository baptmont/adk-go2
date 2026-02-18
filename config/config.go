// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/loopagent"
	"google.golang.org/adk/agent/workflowagents/parallelagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/internal/llminternal/googlellm"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/tool"
)

// CodeConfig represents a reference to a function or callback.
// Equivalent to: common_configs.CodeConfig
type CodeConfig struct {
	// Name of the function/method (e.g., "my_pkg.security.Check")
	Name string `yaml:"name"`

	// Optional params if your system supports parameterized callbacks
	Params map[string]any `yaml:"params,omitempty"`
}

// AgentRefConfig represents a reference to a sub-agent.
// Equivalent to: common_configs.AgentRefConfig
type AgentRefConfig struct {
	// Path to another agent's YAML file
	ConfigPath string `yaml:"config_path,omitempty"`

	// OR an inline code reference
	Code string `yaml:"code,omitempty"`
}

type ToolConfig struct {
	// Name of the tool/method (e.g., "my_pkg.security.Check")
	Name string `yaml:"name"`

	// Optional params if your system supports parameterized callbacks
	Args map[string]any `yaml:"args,omitempty"`
}

// BaseAgentConfig matches the Python BaseAgentConfig Pydantic model.
//
// Usage: Do not use this struct directly for unmarshalling specific agents.
// Embed it into concrete agent configs (see Example below).
type BaseAgentConfig struct {
	// Required. The class of the agent.
	// Default is "BaseAgent" in Python, but usually overridden by concrete agents.
	AgentClass string `yaml:"agent_class"`

	// Required. The name of the agent.
	Name string `yaml:"name"`

	// Optional. Description of the agent.
	Description string `yaml:"description,omitempty"`

	// Optional. List of sub-agents.
	SubAgents []AgentRefConfig `yaml:"sub_agents,omitempty"`

	// Optional. Callbacks to run before execution.
	BeforeAgentCallbacks []CodeConfig `yaml:"before_agent_callbacks,omitempty"`

	// Optional. Callbacks to run after execution.
	AfterAgentCallbacks []CodeConfig `yaml:"after_agent_callbacks,omitempty"`

	// Path to the config file.
	ConfigPath string `yaml:"-"`

	// Handle extra fields (extra='allow'):
	// If you use this struct standalone, this map catches unknown fields.
	// However, the preferred pattern is to embed this struct in a concrete config
	// so specific fields are strongly typed.
	AdditionalProperties map[string]any `yaml:",inline"`
}

// LLMAgentYAMLConfig is the concrete config for a specific agent.
type LLMAgentYAMLConfig struct {
	// 1. Embed BaseAgentConfig with ",inline".
	// This pulls "name", "sub_agents", etc. to the top level of the YAML.
	BaseAgentConfig `yaml:",inline"`

	// 2. Define the "extra" fields specific to this agent here.
	Model string `yaml:"model"`

	Instruction string `yaml:"instruction"`

	Tools []ToolConfig `yaml:"tools,omitempty"`

	GenerateContentConfig *genai.GenerateContentConfig `yaml:"generate_content_config,omitempty"`
}

func (c *LLMAgentYAMLConfig) ToLLMAgentConfig(ctx context.Context) (*llmagent.Config, error) {
	if !googlellm.IsGeminiModel(c.Model) {
		return nil, fmt.Errorf("model %s is not supported", c.Model)
	}

	model, err := gemini.NewModel(ctx, c.Model, &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	subAgents, err := resolveSubAgents(ctx, c.ConfigPath, c.SubAgents)
	if err != nil {
		return nil, err
	}

	tools, err := resolveTools(ctx, c.ConfigPath, c.Tools)
	if err != nil {
		return nil, err
	}

	return &llmagent.Config{
		Name:        c.Name,
		Description: c.Description,
		SubAgents:   subAgents,
		Model:       model,
		Instruction: c.Instruction,
		Tools:       tools,
		GenerateContentConfig: c.GenerateContentConfig,
	}, nil
}

type LoopAgentYAMLConfig struct {
	BaseAgentConfig `yaml:",inline"`
	MaxIterations   uint `yaml:"max_iterations"`
}

func (c *LoopAgentYAMLConfig) ToLoopAgentConfig(ctx context.Context) (*loopagent.Config, error) {
	subAgents, err := resolveSubAgents(ctx, c.ConfigPath, c.SubAgents)
	if err != nil {
		return nil, err
	}

	return &loopagent.Config{
		AgentConfig: agent.Config{
			Name:        c.Name,
			Description: c.Description,
			SubAgents:   subAgents,
		},
		MaxIterations: c.MaxIterations,
	}, nil
}

// ParallelAgentYAMLConfig is the concrete config for a specific agent.
type ParallelAgentYAMLConfig struct {
	BaseAgentConfig `yaml:",inline"`
}

func (c *ParallelAgentYAMLConfig) ToParallelAgentConfig(ctx context.Context) (*parallelagent.Config, error) {
	subAgents, err := resolveSubAgents(ctx, c.ConfigPath, c.SubAgents)
	if err != nil {
		return nil, err
	}

	return &parallelagent.Config{
		AgentConfig: agent.Config{
			Name:        c.Name,
			Description: c.Description,
			SubAgents:   subAgents,
		},
	}, nil
}

// SequentialAgentYAMLConfig is the concrete config for a specific agent.
type SequentialAgentYAMLConfig struct {
	BaseAgentConfig `yaml:",inline"`
}

func (c *SequentialAgentYAMLConfig) ToSequentialAgentConfig(ctx context.Context) (*sequentialagent.Config, error) {
	subAgents, err := resolveSubAgents(ctx, c.ConfigPath, c.SubAgents)
	if err != nil {
		return nil, err
	}

	return &sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        c.Name,
			Description: c.Description,
			SubAgents:   subAgents,
		},
	}, nil
}

func resolveSubAgents(ctx context.Context, parentPath string, refs []AgentRefConfig) ([]agent.Agent, error) {
	var agents []agent.Agent
	for _, ref := range refs {
		if ref.ConfigPath != "" {
			a, err := ResolveAgentReference(ctx, parentPath, ref.ConfigPath)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve agent reference %s: %w", ref.ConfigPath, err)
			}
			agents = append(agents, a)
		} else if ref.Code != "" {
			return nil, fmt.Errorf("inline code agent references are not yet supported for %s", ref.Code)
		}
	}
	return agents, nil
}

func resolveTools(ctx context.Context, parentPath string, toolConfigs []ToolConfig) ([]tool.Tool, error) {
	var tools []tool.Tool
	for _, tc := range toolConfigs {
		if tc.Name != "" {
			ctx = context.WithValue(ctx, "parentPath", parentPath)
			a, err := ResolveToolReference(ctx, tc.Name, tc.Args)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve tool reference %s: %w", tc.Name, err)
			}
			tools = append(tools, a)
		}
	}
	return tools, nil
}
