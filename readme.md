# OpenRouter API Go Client

This library provides a Go client for interacting with the OpenRouter API. It allows you to easily send chat completion requests and receive responses, both synchronously and via streaming.

## Installation

```bash
go get github.com/wojtess/openrouter-api-go
```

## Usage

### Synchronous Request

```go
package main

import (
	"fmt"
	"github.com/wojtess/openrouter-api-go"
)

func main() {
	client := openrouterapigo.NewOpenRouterClient("YOUR_OPENROUTER_API_KEY")

	request := openrouterapigo.Request{
		Messages: []openrouterapigo.MessageRequest{
			{Role: openrouterapigo.RoleUser, Content: "Hello"},
		},
	}

	response, err := client.FetchChatCompletions(request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", response.Choices[0].Message.Content)
}

```

### Streaming Request

```go
package main

import (
	"context"
	"fmt"
	"github.com/wojtesss/openrouter-api-go"
)

func main() {
	client := openrouterapigo.NewOpenRouterClient("YOUR_OPENROUTER_API_KEY")

	request := openrouterapigo.Request{
		Model: "meta-llama/llama-3.2-1b-instruct",
		Messages: []openrouterapigo.MessageRequest{
			{Role: openrouterapigo.RoleUser, Content: "Hello"},
		},
		Stream: true, // Enable streaming
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
			fmt.Printf("%s", output.Choices[0].Delta.Content) // Access delta content for streaming responses
		case <-processingChan:
			// Handle processing events (optional)
		case err := <-errChan:
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			return
		case <-ctx.Done():
			fmt.Println("Context cancelled")
			return
		}
	}
}

```

### Router Agent

The `router_agent.go` file introduces a `RouterAgent`.  The `RouterAgent` simplifies the API for processing requests, abstracting away the need to manage channels and context directly for streaming requests.

#### RouterAgent Example

```go
client := openrouterapigo.NewOpenRouterClient("YOUR_OPENROUTER_API_KEY")
agent := openrouterapigo.NewRouterAgent(client, "your-model", openrouterapigo.RouterAgentConfig{})
response, err := agent.Completion("your prompt")
// or for streaming
agent.CompletionStream("your prompt", outputChan, processingChan, errChan, ctx)
```

#### RouterAgentChat Example
```go
client := openrouterapigo.NewOpenRouterClient("YOUR_OPENROUTER_API_KEY")
agent := openrouterapigo.NewRouterAgentChat(client, "your-model", openrouterapigo.RouterAgentConfig{}, "Initial system prompt")
agent.Chat("First message")
agent.Chat("Second message")
// Access the conversation history via agent.Messages
```

### Specifying Model

You can specify a specific model to use with the `Model` field in the `Request` struct.  If no model is specified, OpenRouter will select a default model.

```go
request := openrouterapigo.Request{
    Model: "google/flan-t5-xxl",
    Messages: []openrouterapigo.MessageRequest{
        {Role: openrouterapigo.RoleUser, Content: "Translate 'Hello' to French."},
    },
}
```

### Setting Provider Preferences

You can set provider preferences using the `Provider` field in the `Request` struct. This allows you to specify the `RefererURL` and `SiteName` for your request.

```go
request := openrouterapigo.Request{
    // ... other request fields
    Provider: &openrouterapigo.ProviderPreferences{
        RefererURL: "https://yourwebsite.com",
        SiteName:   "Your Website Name",
    },
}
```


## Contributing

Contributions are welcome! Feel free to open issues and submit pull requests.


## License

This project is licensed under the MIT License
