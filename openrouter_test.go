package openrouterapigo_test

import (
	"context"
	"fmt"
	"testing"

	openrouterapigo "github.com/wojtess/openrouter-api-go"
)

func TestFetchChatCompletions(t *testing.T) {
	client := openrouterapigo.NewOpenRouterClient("YOUR_TOKEN")

	request := openrouterapigo.Request{
		Model: "meta-llama/llama-3.2-1b-instruct",
		Messages: []openrouterapigo.MessageRequest{
			{openrouterapigo.RoleUser, "Hi", "", ""},
		},
	}

	output, err := client.FetchChatCompletions(request)
	if err != nil {
		t.Errorf("error %v", err)
	}

	t.Logf("output: %v", output.Choices[0].Message.Content)
}

func TestFetchChatCompletionsStreaming(t *testing.T) {
	client := openrouterapigo.NewOpenRouterClient("YOUR_TOKEN")

	request := openrouterapigo.Request{
		Model: "meta-llama/llama-3.2-1b-instruct",
		Messages: []openrouterapigo.MessageRequest{
			{openrouterapigo.RoleUser, "Hello", "", ""},
		},
		Stream: true,
	}

	outputChan := make(chan openrouterapigo.Response)
	processingChan := make(chan interface{})
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go client.FetchChatCompletionsStream(request, outputChan, processingChan, errChan, ctx)

	for {
		select {
		case output := <-outputChan:
			if len(output.Choices) > 0 {
				t.Logf("%s", output.Choices[0].Delta.Content)
			}
		case <-processingChan:
			t.Logf("Processing\n")
		case err := <-errChan:
			if err != nil {
				t.Errorf("Error: %v", err)
				return
			}
			return
		case <-ctx.Done():
			fmt.Println("Context cancelled:", ctx.Err())
			return
		}
	}

}
