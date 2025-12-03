package openrouterapigo

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFirstChoiceMessageErrors(t *testing.T) {
	tests := []struct {
		name     string
		response *Response
	}{
		{
			name:     "nil response",
			response: nil,
		},
		{
			name: "empty choices",
			response: &Response{
				Choices: []Choice{},
			},
		},
		{
			name: "nil message in first choice",
			response: &Response{
				Choices: []Choice{
					{Message: nil},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := firstChoiceMessage(tt.response)
			if err == nil {
				t.Fatalf("expected error, got nil (message=%v)", msg)
			}
		})
	}
}

func TestGenerateMessagesForRequest_OmitsContentWhenToolCalls(t *testing.T) {
	toolID := "tool-1"
	msg := &MessageResponse{
		Role:    RoleAssistant,
		Content: "",
		ToolCalls: []ToolCall{
			{
				ID:   toolID,
				Type: "function",
				Function: ToolCallFunction{
					Name:      "dummy",
					Arguments: "{}",
				},
			},
		},
	}

	msgs := generateMessagesForRequest([]message{msg})
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	data, err := json.Marshal(msgs[0])
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, `"content"`) {
		t.Fatalf("content field should be omitted, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"tool_calls"`) {
		t.Fatalf("tool_calls must be present, got: %s", jsonStr)
	}
}
