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

package llminternal_test

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/genai"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/internal/httprr"
	"google.golang.org/adk/internal/testutil"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

type SumArgs struct {
	A int `json:"a"` // an integer to sum
	B int `json:"b"` // another integer to sum
}
type SumResult struct {
	Sum int `json:"sum"` // the sum of two integers
}

func sumFunc(ctx tool.Context, input SumArgs) (SumResult, error) {
	return SumResult{Sum: input.A + input.B}, nil
}

func TestParallelFunctionCalls(t *testing.T) {
	httpRecordFilename := filepath.Join("testdata", strings.ReplaceAll(t.Name(), "/", "_")+".httprr")

	baseTransport, err := testutil.NewGeminiTransport(httpRecordFilename)
	if err != nil {
		t.Fatal(err)
	}

	apiKey := ""
	if recording, _ := httprr.Recording(httpRecordFilename); !recording {
		apiKey = "fakekey"
	}

	cfg := &genai.ClientConfig{
		HTTPClient: &http.Client{Transport: baseTransport},
		APIKey:     apiKey,
	}

	geminiModel, err := gemini.NewModel(t.Context(), "gemini-3-pro-preview", cfg)
	if err != nil {
		t.Fatal(err)
	}

	sumTool, err := functiontool.New(functiontool.Config{
		Name:        "sum",
		Description: "sums two integers",
	}, sumFunc)
	if err != nil {
		t.Fatal(err)
	}

	a, err := llmagent.New(llmagent.Config{
		Name:        "calculator",
		Description: "A calculator that can add two integers",
		Instruction: "You are a calculator assistant. You will recieve requests to add two integers. Respond with the sum of the two integers and you must use the sum tool to calculate the sum.",
		Model:       geminiModel,
		Tools: []tool.Tool{
			sumTool,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	sessionService := session.InMemoryService()
	_, err = sessionService.Create(t.Context(), &session.CreateRequest{
		AppName:   "testApp",
		UserID:    "testUser",
		SessionID: "testSession",
	})
	if err != nil {
		t.Fatal(err)
	}

	r, err := runner.New(runner.Config{
		Agent:          a,
		SessionService: sessionService,
		AppName:        "testApp",
	})
	if err != nil {
		t.Fatal(err)
	}

	it := r.Run(t.Context(), "testUser", "testSession", &genai.Content{
		Parts: []*genai.Part{
			genai.NewPartFromText("Can you add 2 and 3? Also 4 and 5? And 6 and 7?"),
		},
		Role: "user",
	}, agent.RunConfig{StreamingMode: agent.StreamingModeSSE})
	for _, err := range it {
		if err != nil {
			t.Fatal(err)
		}
	}
}
