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

package llminternal

import (
	"fmt"
	"iter"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/internal/utils"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
)

// identityRequestProcessor populates the LLMRequest's Identity based on
// the InvocationContext that includes the previous events.
func identityRequestProcessor(ctx agent.InvocationContext, req *model.LLMRequest, f *Flow) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		agent := ctx.Agent()
		si := fmt.Sprintf("You are an agent. Your internal name is %q.", agent.Name())
		if agent.Description() != "" {
			si += fmt.Sprintf(" The description about you is %q.", agent.Description())
		}
		utils.AppendInstructions(req, si)
	}
}
