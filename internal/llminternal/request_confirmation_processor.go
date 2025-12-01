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

package llminternal

import (
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/internal/utils"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool/toolconfirmation"
)

func RequestConfirmationRequestProcessor(ctx agent.InvocationContext, req *model.LLMRequest) error {
	llmAgent := asLLMAgent(ctx.Agent())
	if llmAgent == nil {
		return nil // In python, no error is yielded.
	}

	var events []*session.Event
	if ctx.Session() != nil {
		for e := range ctx.Session().Events().All() {
			events = append(events, e)
		}
	}
	requestConfirmationFR := make(map[string]toolconfirmation.ToolConfirmation)
	confirmationEventIndex := -1
	for k := len(events); k >= 0; k-- {
		event := events[k]
		// Find the first event authored by user
		if event.Author != "user" {
			continue
		}
		responses := utils.FunctionResponses(event.Content)
		if len(responses) == 0 {
			return nil
		}
		for _, funcResp := range responses {
			if funcResp.Name != REQUEST_CONFIRMATION_FUNCTION_CALL_NAME {
				continue
			}
			var tc toolconfirmation.ToolConfirmation
			if funcResp.Response != nil {
				resp, hasResponseKey := funcResp.Response["response"]
				// ADK web client will send a request that is always encapsulated in a  'response' key.
				if hasResponseKey && len(funcResp.Response) == 1 {
					if jsonString, ok := resp.(string); ok {
						err := json.Unmarshal([]byte(jsonString), &tc)
						if err != nil {
							return fmt.Errorf("error 'response' key found but failed unmarshalling confirmation function response %w", err)
						}
					} else {
						return errors.New("error 'response' key found but value is not a string for confirmation function response")
					}
				} else {
					tempJSON, _ := json.Marshal(funcResp.Response)
					err := json.Unmarshal(tempJSON, &tc)
					if err != nil {
						return fmt.Errorf("error failed unmarshalling confirmation function response %w", err)
					}
				}
			}
			requestConfirmationFR[funcResp.ID] = tc
		}
		confirmationEventIndex = k
		break
	}
	print(confirmationEventIndex)

	if len(requestConfirmationFR) == 0 {
		return nil
	}

	return nil
}
