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
	LlmRequest *model.LLMRequest `yaml:"llm_request,omitempty"`

	// Required. The LLM response.
	LlmResponse *LLMResponseRecording `yaml:"llm_response,omitempty"`
}

type LLMResponseRecording struct {
	Content           *genai.Content `yaml:"content,omitempty"`
	UsageMetadata     *localUsageMetadata `yaml:"usage_metadata,omitempty"`
	LogprobsResult    *genai.LogprobsResult `yaml:"logprobs_result,omitempty"`
	Partial           bool `yaml:"partial,omitempty"`
	TurnComplete      bool `yaml:"turn_complete,omitempty"`
	Interrupted       bool `yaml:"interrupted,omitempty"`
	ErrorCode         string `yaml:"error_code,omitempty"`
	ErrorMessage      string `yaml:"error_message,omitempty"`
	FinishReason      genai.FinishReason `yaml:"finish_reason,omitempty"`
	AvgLogprobs       float64 `yaml:"avg_logprobs,omitempty"`
}

func (l *LLMResponseRecording) ToLLMResponse() *model.LLMResponse {
	return &model.LLMResponse{
		Content:           l.Content,
		UsageMetadata:     l.UsageMetadata.ToGenAI(),
		LogprobsResult:    l.LogprobsResult,
		Partial:           l.Partial,
		TurnComplete:      l.TurnComplete,
		Interrupted:       l.Interrupted,
		ErrorCode:         l.ErrorCode,
		ErrorMessage:      l.ErrorMessage,
		FinishReason:      l.FinishReason,
		AvgLogprobs:       l.AvgLogprobs,
	}
}

type localUsageMetadata struct {
	CacheTokensDetails []*localModalityTokenCount `yaml:"cache_tokens_details,omitempty"`
	CachedContentTokenCount int32 `yaml:"cached_content_token_count,omitempty"`
	CandidatesTokenCount int32 `yaml:"candidates_token_count,omitempty"`
	CandidatesTokensDetails []*localModalityTokenCount `yaml:"candidates_tokens_details,omitempty"`
	PromptTokenCount int32 `yaml:"prompt_token_count,omitempty"`
	PromptTokensDetails []*localModalityTokenCount `yaml:"prompt_tokens_details,omitempty"`
	ThoughtsTokenCount int32 `yaml:"thoughts_token_count,omitempty"`
	ToolUsePromptTokenCount int32 `yaml:"tool_use_prompt_token_count,omitempty"`
	ToolUsePromptTokensDetails []*localModalityTokenCount `yaml:"tool_use_prompt_tokens_details,omitempty"`
	TotalTokenCount int32 `yaml:"total_token_count,omitempty"`
	TrafficType string `yaml:"traffic_type,omitempty"`
}

func (l *localUsageMetadata) ToGenAI() *genai.GenerateContentResponseUsageMetadata {
	if l == nil {
			return nil
	}

	return &genai.GenerateContentResponseUsageMetadata{
        CacheTokensDetails: transformModalityTokenCount(l.CacheTokensDetails),
        CachedContentTokenCount: l.CachedContentTokenCount,
        CandidatesTokenCount: l.CandidatesTokenCount,
        CandidatesTokensDetails: transformModalityTokenCount(l.CandidatesTokensDetails),
        PromptTokenCount: l.PromptTokenCount,
        PromptTokensDetails: transformModalityTokenCount(l.PromptTokensDetails),
        ThoughtsTokenCount: l.ThoughtsTokenCount,
        ToolUsePromptTokenCount: l.ToolUsePromptTokenCount,
        ToolUsePromptTokensDetails: transformModalityTokenCount(l.ToolUsePromptTokensDetails),
        TotalTokenCount: l.TotalTokenCount,
        TrafficType: genai.TrafficType(l.TrafficType),
    }
}

func transformModalityTokenCount(l []*localModalityTokenCount) []*genai.ModalityTokenCount {
	if l == nil {
		return nil
	}
	var result []*genai.ModalityTokenCount
	for _, item := range l {
		result = append(result, item.ToGenAI())
	}
	return result
}

func (l *localModalityTokenCount) ToGenAI() *genai.ModalityTokenCount {
	if l == nil {
		return nil
	}
	return &genai.ModalityTokenCount{
		Modality: l.Modality,
		TokenCount: l.TokenCount,
	}
}

type localModalityTokenCount struct {
    Modality genai.MediaModality `yaml:"modality,omitempty"`
    TokenCount int32 `yaml:"token_count,omitempty"`
}

// ToolRecording represents a paired tool call and response.
type ToolRecording struct {
	// Required. The tool call.
	ToolCall *genai.FunctionCall `yaml:"tool_call,omitempty"`

	// Required. The tool response.
	ToolResponse *genai.FunctionResponse `yaml:"tool_response,omitempty"`
}

// Recording represents a single interaction recording, ordered by request timestamp.
type Recording struct {
	// Index of the user message this recording belongs to (0-based).
	UserMessageIndex int `yaml:"user_message_index"`

	// Name of the agent.
	AgentName string `yaml:"agent_name"`

	// oneof fields - start

	// LLM request-response pair.
	LLMRecording *LLMRecording `yaml:"llm_recording,omitempty"`

	// Tool call-response pair.
	ToolRecording *ToolRecording `yaml:"tool_recording,omitempty"`

	// oneof fields - end
}

// Recordings represents all recordings in chronological order.
type Recordings struct {
	// Chronological list of all recordings.
	Recordings []Recording `yaml:"recordings"`
}
