package openrouterapigo

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestChoiceSelector_Default(t *testing.T) {
	selector := defaultChoiceSelector

	_, err := selector([]Choice{})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}

	msg := &MessageResponse{Content: "hi"}
	ch, err := selector([]Choice{{Message: msg}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.Message != msg {
		t.Fatalf("expected first choice to be selected")
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
