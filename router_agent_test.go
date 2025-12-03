package openrouterapigo

import "testing"

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

