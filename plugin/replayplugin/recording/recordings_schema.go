/*
 * Copyright 2026 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package recording

import (
	"google.golang.org/genai"

	"google.golang.org/adk/model"
)

// LLMRecording represents a paired LLM request and response.
type LLMRecording struct {
	// Required. The LLM request.
	LlmRequest *model.LLMRequest `json:"llm_request,omitempty" yaml:"llm_request,omitempty"`

	// Required. The LLM response.
	LlmResponse *model.LLMResponse `json:"llm_response,omitempty" yaml:"llm_response,omitempty"`
}

// ToolRecording represents a paired tool call and response.
type ToolRecording struct {
	// Required. The tool call.
	ToolCall *genai.FunctionCall `json:"tool_call,omitempty" yaml:"tool_call,omitempty"`

	// Required. The tool response.
	ToolResponse *genai.FunctionResponse `json:"tool_response,omitempty" yaml:"tool_response,omitempty"`
}

// Recording represents a single interaction recording, ordered by request timestamp.
type Recording struct {
	// Index of the user message this recording belongs to (0-based).
	UserMessageIndex int `json:"user_message_index" yaml:"user_message_index"`

	// Name of the agent.
	AgentName string `json:"agent_name" yaml:"agent_name"`

	// oneof fields - start

	// LLM request-response pair.
	LLMRecording *LLMRecording `json:"llm_recording,omitempty" yaml:"llm_recording,omitempty"`

	// Tool call-response pair.
	ToolRecording *ToolRecording `json:"tool_recording,omitempty" yaml:"tool_recording,omitempty"`

	// oneof fields - end
}

// Recordings represents all recordings in chronological order.
type Recordings struct {
	// Chronological list of all recordings.
	Recordings []Recording `json:"recordings" yaml:"recordings"`
}
