// Copyright 2025 Google LLC
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

package models

import (
	"google.golang.org/genai"

	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
)

// EventActions represent a data model for session.EventActions
type EventActions struct {
	StateDelta    map[string]any   `json:"stateDelta"`
	ArtifactDelta map[string]int64 `json:"artifactDelta"`
}

// Event represents a single event in a session.
type Event struct {
	ID                 string                   `json:"id"`
	InvocationID       string                   `json:"invocationId"`
	Branch             string                   `json:"branch,omitempty"`
	Author             string                   `json:"author"`
	Partial            bool                     `json:"partial,omitempty"`
	LongRunningToolIDs []string                 `json:"longRunningToolIds,omitempty"`
	Content            *genai.Content           `json:"content"`
	GroundingMetadata  *genai.GroundingMetadata `json:"groundingMetadata"`
	UsageMetadata      *genai.GenerateContentResponseUsageMetadata `json:"usageMetadata"`
	TurnComplete       bool                     `json:"turnComplete,omitempty"`
	Interrupted        bool                     `json:"interrupted,omitempty"`
	ErrorCode          string                   `json:"errorCode,omitempty"`
	ErrorMessage       string                   `json:"errorMessage,omitempty"`
	AvgLogprobs        float64                  `json:"avgLogprobs,omitempty"`
	FinishReason       genai.FinishReason       `json:"finishReason,omitempty"`
	Actions            EventActions             `json:"actions"`
}

// ToSessionEvent maps Event data struct to session.Event
func ToSessionEvent(event Event) *session.Event {
	return &session.Event{
		ID:                 event.ID,
		InvocationID:       event.InvocationID,
		Branch:             event.Branch,
		Author:             event.Author,
		LongRunningToolIDs: event.LongRunningToolIDs,
		LLMResponse: model.LLMResponse{
			AvgLogprobs:       event.AvgLogprobs,
			Content:           event.Content,
			GroundingMetadata: event.GroundingMetadata,
			UsageMetadata:     event.UsageMetadata,
			Partial:           event.Partial,
			TurnComplete:      event.TurnComplete,
			Interrupted:       event.Interrupted,
			ErrorCode:         event.ErrorCode,
			ErrorMessage:      event.ErrorMessage,
			FinishReason:      event.FinishReason,
		},
		Actions: session.EventActions{
			StateDelta:    event.Actions.StateDelta,
			ArtifactDelta: event.Actions.ArtifactDelta,
		},
	}
}

// FromSessionEvent maps session.Event to Event data struct
func FromSessionEvent(event session.Event) Event {
	return Event{
		ID:                 event.ID,
		InvocationID:       event.InvocationID,
		Branch:             event.Branch,
		Author:             event.Author,
		Partial:            event.Partial,
		LongRunningToolIDs: event.LongRunningToolIDs,
		AvgLogprobs:        event.LLMResponse.AvgLogprobs,
		Content:            event.LLMResponse.Content,
		GroundingMetadata:  event.LLMResponse.GroundingMetadata,
		UsageMetadata:      event.LLMResponse.UsageMetadata,
		TurnComplete:       event.LLMResponse.TurnComplete,
		Interrupted:        event.LLMResponse.Interrupted,
		ErrorCode:          event.LLMResponse.ErrorCode,
		ErrorMessage:       event.LLMResponse.ErrorMessage,
		FinishReason:       event.LLMResponse.FinishReason,
		Actions: EventActions{
			StateDelta:    event.Actions.StateDelta,
			ArtifactDelta: event.Actions.ArtifactDelta,
		},
	}
}
