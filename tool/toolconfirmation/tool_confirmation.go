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

// Package toolconfirmation provides structures and utilities for handling
// Human-in-the-Loop tool execution confirmations within the ADK.
package toolconfirmation

// ToolConfirmation represents the state and details of a user confirmation request
// for a tool execution.
type ToolConfirmation struct {
	// Hint is the message provided to the user to explain why the confirmation
	// is needed and what action is being confirmed.
	Hint string

	// Confirmed indicates the user's decision.
	// true if the user approved the action, false if they denied it.
	// The state before the user has responded is typically handled outside
	// this struct (e.g., by the absence of a result or a pending status).
	Confirmed bool

	// Payload contains any additional data or context related to the confirmation request.
	// The structure of the Payload is application-specific.
	Payload any
}
