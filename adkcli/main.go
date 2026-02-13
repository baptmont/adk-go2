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
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"google.golang.org/genai"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/config"
	"google.golang.org/adk/plugin"
	"google.golang.org/adk/plugin/replayplugin"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/tool"
)

func main() {
	// 1. Get the Current Working Directory (where the user typed 'adk')
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	// 2. Look for the configuration file
	configPath := filepath.Join(cwd, "root_agent.yaml")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("❌ No 'root_agent.yaml' found in %s\n", cwd)
		fmt.Println("Usage: Go to a folder with a 'root_agent.yaml' and run 'adk'")
		os.Exit(1)
	}

	fmt.Printf("🚀 Found agent config in: %s\n", cwd)

	// 3. Load the agent using the Factory we built earlier
	// This reads the YAML, finds the 'agent_class', and calls the registered factory.
	myAgent, err := config.FromConfig(context.Background(), configPath)
	if err != nil {
		log.Fatalf("Error loading agent: %v", err)
	}

	fmt.Printf("✅ Agent loaded successfully: %s\n", myAgent.Name())

	ctx := context.Background()

	p, err := plugin.New(plugin.Config{
		Name: "test",
		BeforeAgentCallback: func(ctx agent.CallbackContext) (*genai.Content, error) {
			fmt.Printf("\n🤖 BeforeAgentCallback for %s\n", ctx.AgentName())
			return nil, nil
		},
		AfterToolCallback: func(ctx tool.Context, tool tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
			fmt.Printf("\n🤖 AfterToolCallback for %s call to %s\n", ctx.AgentName(), tool.Name())
			return nil, nil
		},
	})
	if err != nil {
		log.Fatalf("Error loading plugin: %v", err)
	}

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(myAgent),
		PluginConfig: runner.PluginConfig{
			Plugins: []*plugin.Plugin{
				p,
				replayplugin.MustNew(),
			},
		},
	}

	l := full.NewLauncher()
	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
