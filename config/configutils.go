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

// configutils.go provides utility functions for working with configurable agents.
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/loopagent"
	"google.golang.org/adk/agent/workflowagents/parallelagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/agenttool"
	"google.golang.org/adk/tool/exitlooptool"
	"google.golang.org/adk/tool/geminitool"
)

type AgentFactory func(ctx context.Context, configBytes []byte, configPath string) (agent.Agent, error)

type ToolFactory func(ctx context.Context, args map[string]any) (tool.Tool, error)

var (
	registryMu    sync.RWMutex
	registry      = make(map[string]AgentFactory)
	agentRegistry = make(map[string]agent.Agent)
	toolRegistry  = make(map[string]ToolFactory)
)

func init() {
	Register("LlmAgent", NewLLMAgent)
	Register("LoopAgent", NewLoopAgent)
	Register("ParallelAgent", NewParallelAgent)
	Register("SequentialAgent", NewSequentialAgent)
	RegisterTool("exit_loop", func(ctx context.Context, _ map[string]any) (tool.Tool, error) {
		return exitlooptool.New()
	})
	RegisterTool("google_search", func(ctx context.Context, _ map[string]any) (tool.Tool, error) {
		return geminitool.GoogleSearch{}, nil
	})
	RegisterTool("AgentTool", func(ctx context.Context, args map[string]any) (tool.Tool, error) {
		if args == nil {
			return nil, fmt.Errorf("args is nil")
		}
		a, ok := args["agent"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("agent not found in args")
		}
		skipSummarization := false
		if ss, ok := a["skip_summarization"].(bool); ok {
			skipSummarization = ss
		}
		parentPath := ctx.Value("parentPath").(string)
		if configPath, ok := a["config_path"].(string); ok {
			ag, err := ResolveAgentReference(ctx, parentPath, configPath)
			if err != nil {
				return nil, err
			}
			return agenttool.New(ag, &agenttool.Config{SkipSummarization: skipSummarization}), nil
		} else {
			return nil, fmt.Errorf("config_path not found in args")
		}
	})
}

// Register allows concrete implementations to add themselves to the system.
// This replaces Python's dynamic importlib logic.
func Register(name string, factory AgentFactory) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := registry[name]; dup {
		return fmt.Errorf("Register called twice for agent %s", name)
	}
	registry[name] = factory
	return nil
}

// RegisterTool allows concrete implementations to add themselves to the system.
func RegisterTool(name string, factory ToolFactory) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := toolRegistry[name]; dup {
		return fmt.Errorf("RegisterTool called twice for tool %s", name)
	}
	toolRegistry[name] = factory
	return nil
}

// FromConfig builds an agent from a config file path.
// Equivalent to: def from_config(config_path: str) -> BaseAgent
func FromConfig(ctx context.Context, configPath string) (agent.Agent, error) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// 1. Read the file
	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", absPath)
		}
		return nil, err
	}

	// 2. Peek at the "agent_class" field to know which factory to use.
	var baseConfig BaseAgentConfig
	if err := yaml.Unmarshal(data, &baseConfig); err != nil {
		return nil, fmt.Errorf("invalid YAML content: %w", err)
	}

	// Default fallback similar to Python's handling
	agentClass := baseConfig.AgentClass
	if agentClass == "" {
		agentClass = "LlmAgent"
	}

	// 3. Resolve the factory (The Go equivalent of _resolve_agent_class)
	registryMu.RLock()
	factory, exists := registry[agentClass]
	registryMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("invalid agent class '%s': not registered. Ensure the package is imported", agentClass)
	}

	// 4. Delegate creation to the specific factory.
	// We pass the raw data so the factory can unmarshal into its specific Config struct.
	return factory(ctx, data, absPath)
}

func ResolveToolReference(ctx context.Context, toolName string, args map[string]any) (tool.Tool, error) {
	if toolName == "" {
		return nil, fmt.Errorf("tool name cannot be empty")
	}

	registryMu.RLock()
	if t, ok := toolRegistry[toolName]; ok {
		registryMu.RUnlock()
		return t(ctx, args)
	}
	registryMu.RUnlock()
	return nil, fmt.Errorf("tool '%s' not found", toolName)
}

// ResolveAgentReference builds an agent from a reference config.
func ResolveAgentReference(ctx context.Context, parentPath string, refPath string) (agent.Agent, error) {
	if refPath == "" {
		return nil, fmt.Errorf("agent reference path cannot be empty")
	}

	targetPath := refPath
	// Handle relative paths
	if !filepath.IsAbs(refPath) {
		targetPath = filepath.Join(filepath.Dir(parentPath), refPath)
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	registryMu.RLock()
	if a, ok := agentRegistry[absPath]; ok {
		registryMu.RUnlock()
		return a, nil
	}
	registryMu.RUnlock()

	a, err := FromConfig(ctx, absPath)
	if err != nil {
		return nil, err
	}

	registryMu.Lock()
	defer registryMu.Unlock()
	if existing, ok := agentRegistry[absPath]; ok {
		return existing, nil
	}
	agentRegistry[absPath] = a
	return a, nil
}

// NewLLMAgent is the factory function registered in the system.
func NewLLMAgent(ctx context.Context, data []byte, configPath string) (agent.Agent, error) {
	var cfg LLMAgentYAMLConfig

	// Unmarshal parses the shared fields (Name) into BaseAgentConfig
	// AND the specific fields (ModelName) into LLMAgentConfig simultaneously.
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse LLM agent config: %w", err)
	}

	// Validation Logic (Pydantic equivalent)
	if cfg.Name == "" {
		return nil, fmt.Errorf("'name' is required")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("'model' is required for LlmAgent")
	}

	cfg.ConfigPath = configPath

	agentConfig, err := cfg.ToLLMAgentConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM agent config: %w", err)
	}

	return llmagent.New(*agentConfig)
}

func NewLoopAgent(ctx context.Context, data []byte, configPath string) (agent.Agent, error) {
	var cfg LoopAgentYAMLConfig

	// Unmarshal parses the shared fields (Name) into BaseAgentConfig
	// AND the specific fields (ModelName) into LLMAgentConfig simultaneously.
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse Loop agent config: %w", err)
	}

	// Validation Logic (Pydantic equivalent)
	if cfg.Name == "" {
		return nil, fmt.Errorf("'name' is required")
	}

	cfg.ConfigPath = configPath

	agentConfig, err := cfg.ToLoopAgentConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Loop agent config: %w", err)
	}

	return loopagent.New(*agentConfig)
}

func NewParallelAgent(ctx context.Context, data []byte, configPath string) (agent.Agent, error) {
	var cfg ParallelAgentYAMLConfig

	// Unmarshal parses the shared fields (Name) into BaseAgentConfig
	// AND the specific fields (ModelName) into LLMAgentConfig simultaneously.
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse Parallel agent config: %w", err)
	}

	// Validation Logic (Pydantic equivalent)
	if cfg.Name == "" {
		return nil, fmt.Errorf("'name' is required")
	}

	cfg.ConfigPath = configPath

	agentConfig, err := cfg.ToParallelAgentConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Parallel agent config: %w", err)
	}

	return parallelagent.New(*agentConfig)
}

func NewSequentialAgent(ctx context.Context, data []byte, configPath string) (agent.Agent, error) {
	var cfg SequentialAgentYAMLConfig

	// Unmarshal parses the shared fields (Name) into BaseAgentConfig
	// AND the specific fields (ModelName) into LLMAgentConfig simultaneously.
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse Sequential agent config: %w", err)
	}

	// Validation Logic (Pydantic equivalent)
	if cfg.Name == "" {
		return nil, fmt.Errorf("'name' is required")
	}

	cfg.ConfigPath = configPath

	agentConfig, err := cfg.ToSequentialAgentConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Sequential agent config: %w", err)
	}

	return sequentialagent.New(*agentConfig)
}
