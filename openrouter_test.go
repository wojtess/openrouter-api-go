package openrouterapigo_test

import (
	"context"
	"fmt"
	"image"
	"os"
	"path"
	"testing"

	openrouterapigo "github.com/wojtess/openrouter-api-go"
)

func TestFetchChatCompletions(t *testing.T) {
	client := openrouterapigo.NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))

	request := openrouterapigo.Request{
		Model: "meta-llama/llama-3.2-1b-instruct",
		Messages: []openrouterapigo.MessageRequest{
			{openrouterapigo.RoleUser, "Hi", "", ""},
		},
	}

	output, err := client.FetchChatCompletions(request)
	if err != nil {
		t.Errorf("error %v", err)
		return
	}

	t.Logf("output: %v", output.Choices[0].Message.Content)
}

func TestFetchChatCompletionsStreaming(t *testing.T) {
	client := openrouterapigo.NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))

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

func TestFetchChatCompletionsAgentStreaming(t *testing.T) {
	client := openrouterapigo.NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))
	agent := openrouterapigo.NewRouterAgent(client, "meta-llama/llama-3.2-1b-instruct", openrouterapigo.RouterAgentConfig{
		Temperature: 0.7,
		MaxTokens:   100,
	})

	outputChan := make(chan openrouterapigo.Response)
	processingChan := make(chan interface{})
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chat := []openrouterapigo.MessageRequest{
		{Role: openrouterapigo.RoleSystem, Content: "You are a helpful assistant."},
		{Role: openrouterapigo.RoleUser, Content: "Hello"},
	}

	go agent.ChatStream(chat, outputChan, processingChan, errChan, ctx)

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

func TestFetchChatCompletionsAgentSimpleChat(t *testing.T) {
	client := openrouterapigo.NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))
	agent := openrouterapigo.NewRouterAgentChat(client, "meta-llama/llama-3.2-1b-instruct", openrouterapigo.RouterAgentConfig{
		Temperature: 0.0,
		MaxTokens:   100,
	}, "You are helpful asistant, answer in short worlds")

	agent.Chat("Remeber this: \"wojtess\"")
	agent.Chat("What I asked you to rember?")

	for _, msg := range agent.Messages {
		content, ok := msg.Content.(string)
		if ok {
			t.Logf(string(msg.Role) + ": " + string(content))
		}
	}
}

func TestFetchChatCompletionsAgentSimpleChatWithImage(t *testing.T) {
	client := openrouterapigo.NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))
	agent := openrouterapigo.NewRouterAgentChat(client, "google/gemma-3-27b-it" /*Select multimodal model https://openrouter.ai/docs/features/images-and-pdfs*/, openrouterapigo.RouterAgentConfig{
		Temperature: 0.0,
		MaxTokens:   100,
	}, "You are helpful asistant")

	file, err := os.Open(path.Join("data_for_test", "hello_world.png"))
	if err != nil {
		t.Errorf("Error while opening file: %s", err)
		return
	}
	img, _, err := image.Decode(file)
	if err != nil {
		t.Errorf("Error while decoding image: %s", err)
		return
	}

	agent.ChatWithImage("What is in image?", img)

	for _, msg := range agent.Messages {
		content, ok := msg.Content.(string)
		if ok {
			t.Logf(string(msg.Role) + ": " + string(content))
			continue
		}
		content1, ok := msg.Content.([]openrouterapigo.ContentPart)
		if ok {
			//Assumption is that text is on index 0 and image is on index 1..n
			t.Logf(string(msg.Role) + ": " + string(content1[0].Text))
		}
	}
}

func TestFetchChatCompletionsAgentSimpleChatWithPDF(t *testing.T) {
	client := openrouterapigo.NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))
	agent := openrouterapigo.NewRouterAgentChat(client, "google/gemma-3-27b-it" /*Select multimodal model https://openrouter.ai/docs/features/images-and-pdfs*/, openrouterapigo.RouterAgentConfig{
		Temperature: 0.0,
		MaxTokens:   100,
	}, "You are helpful asistant")

	agent.ChatWithPDF("What is in image?", path.Join("data_for_test", "tex_sample.pdf"))

	for _, msg := range agent.Messages {
		content, ok := msg.Content.(string)
		if ok {
			t.Logf(string(msg.Role) + ": " + string(content))
			continue
		}
		content1, ok := msg.Content.([]openrouterapigo.ContentPart)
		if ok {
			//Assumption is that text is on index 0 and pdfs are on index 1..n
			t.Logf(string(msg.Role) + ": " + string(content1[0].Text))
		}
	}
}
