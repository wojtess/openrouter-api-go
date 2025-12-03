package openrouterapigo

import (
	"context"
	"fmt"
	"image"
	"os"
	"path"
	"testing"
)

func TestFetchChatCompletions(t *testing.T) {
	client := NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))

	request := Request{
		Model: "meta-llama/llama-3.2-1b-instruct",
		Messages: []MessageRequest{
			{Role: RoleUser, Content: TextContent("Hi")},
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
	client := NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))

	request := Request{
		Model: "meta-llama/llama-3.2-1b-instruct",
		Messages: []MessageRequest{
			{Role: RoleUser, Content: TextContent("Hello")},
		},
		Stream: true,
	}

	outputChan := make(chan Response)
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
	client := NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))
	agent := NewRouterAgent(client, "meta-llama/llama-3.2-1b-instruct", RouterAgentConfig{
		Temperature: 0.7,
		MaxTokens:   100,
	})

	outputChan := make(chan Response)
	processingChan := make(chan interface{})
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chat := []MessageRequest{
		{Role: RoleSystem, Content: TextContent("You are a helpful assistant.")},
		{Role: RoleUser, Content: TextContent("Hello")},
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
	client := NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))
	agent := NewRouterAgentChat(client, "meta-llama/llama-3.2-1b-instruct", RouterAgentConfig{
		Temperature: 0.0,
		MaxTokens:   100,
	}, "You are helpful asistant, answer in short worlds")

	agent.Chat("Remeber this: \"wojtess\"")
	agent.Chat("What I asked you to rember?")

	for _, msg := range agent.Messages {
		//Assumption is that text is on index 0 and pdfs are on index 1..n
		t.Logf(string(msg.GetRole()) + ": " + string(msg.GetContentPart()[0].Text))
	}
}

func TestFetchChatCompletionsAgentSimpleChatWithImage(t *testing.T) {
	client := NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))
	agent := NewRouterAgentChat(client, "google/gemma-3-27b-it" /*Select multimodal model https://openrouter.ai/docs/features/images-and-pdfs*/, RouterAgentConfig{
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
		//Assumption is that text is on index 0 and pdfs are on index 1..n
		t.Logf(string(msg.GetRole()) + ": " + string(msg.GetContentPart()[0].Text))
	}
}

func TestFetchChatCompletionsAgentSimpleChatWithPDF(t *testing.T) {
	client := NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))
	agent := NewRouterAgentChat(client, "google/gemma-3-27b-it" /*Select multimodal model https://openrouter.ai/docs/features/images-and-pdfs*/, RouterAgentConfig{
		Temperature: 0.0,
		MaxTokens:   100,
	}, "You are helpful asistant")

	agent.ChatWithPDF("What is in image?", path.Join("data_for_test", "tex_sample.pdf"))

	for _, msg := range agent.Messages {
		//Assumption is that text is on index 0 and pdfs are on index 1..n
		t.Logf(string(msg.GetRole()) + ": " + string(msg.GetContentPart()[0].Text))
	}
}

func TestFetchChatCompletionsAgentSimpleChatUsingTool(t *testing.T) {
	client := NewOpenRouterClient(os.Getenv("OPENROUTER_API_KEY"))
	agent := NewRouterAgentChat(client, "mistralai/ministral-8b" /*Select multimodal model https://openrouter.ai/docs/features/images-and-pdfs*/, RouterAgentConfig{
		Temperature: 0.0,
		MaxTokens:   100,
	}, "You are helpful asistant")

	type args struct {
		A int `json:"FirstArgument" desc:"function returns this value"`
	}
	err := AddToolToAgent(&agent, ToolDefinition[args]{
		Function: func(arg args) any {
			return arg.A
		},
		Name:        "test_func",
		Description: "function for testing if function calling is working, use when user ask for use any tool, returns input value",
	})
	if err != nil {
		t.Fatalf("failed to register tool: %v", err)
	}

	_, err = agent.Chat("Use tool")
	if err != nil {
		t.Errorf("error while sending request: %s", err)
	}

	for _, msg := range agent.Messages {
		//Assumption is that text is on index 0 and pdfs are on index 1..n
		if msg.GetRole() == RoleTool {
			t.Logf(string(msg.GetName()) + ": " + string(msg.GetContentPart()[0].Text))
		} else {
			t.Logf(string(msg.GetRole()) + ": " + string(msg.GetContentPart()[0].Text))
		}
	}
}
