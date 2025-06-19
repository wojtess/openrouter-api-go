package openrouterapigo

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
)

type RouterAgentConfig struct {
	ResponseFormat    *ResponseFormat `json:"response_format,omitempty"`
	Stop              []string        `json:"stop,omitempty"`
	MaxTokens         int             `json:"max_tokens,omitempty"`
	Temperature       float64         `json:"temperature,omitempty"`
	Tools             []Tool          `json:"tools,omitempty"`
	ToolChoice        *ToolChoice     `json:"tool_choice,omitempty"`
	Seed              int             `json:"seed,omitempty"`
	TopP              float64         `json:"top_p,omitempty"`
	TopK              int             `json:"top_k,omitempty"`
	FrequencyPenalty  float64         `json:"frequency_penalty,omitempty"`
	PresencePenalty   float64         `json:"presence_penalty,omitempty"`
	RepetitionPenalty float64         `json:"repetition_penalty,omitempty"`
	LogitBias         map[int]float64 `json:"logit_bias,omitempty"`
	TopLogprobs       int             `json:"top_logprobs,omitempty"`
	MinP              float64         `json:"min_p,omitempty"`
	TopA              float64         `json:"top_a,omitempty"`
}

type RouterAgent struct {
	client *OpenRouterClient
	model  string
	config RouterAgentConfig
}

func NewRouterAgent(client *OpenRouterClient, model string, config RouterAgentConfig) *RouterAgent {
	return &RouterAgent{
		client: client,
		model:  model,
		config: config,
	}
}

func (agent RouterAgent) Completion(prompt string) (*Response, error) {
	request := Request{
		Prompt:            prompt,
		Model:             agent.model,
		ResponseFormat:    agent.config.ResponseFormat,
		Stop:              agent.config.Stop,
		MaxTokens:         agent.config.MaxTokens,
		Temperature:       agent.config.Temperature,
		Tools:             agent.config.Tools,
		ToolChoice:        agent.config.ToolChoice,
		Seed:              agent.config.Seed,
		TopP:              agent.config.TopP,
		TopK:              agent.config.TopK,
		FrequencyPenalty:  agent.config.FrequencyPenalty,
		PresencePenalty:   agent.config.PresencePenalty,
		RepetitionPenalty: agent.config.RepetitionPenalty,
		LogitBias:         agent.config.LogitBias,
		TopLogprobs:       agent.config.TopLogprobs,
		MinP:              agent.config.MinP,
		TopA:              agent.config.TopA,
		Stream:            false,
	}

	return agent.client.FetchChatCompletions(request)
}

func (agent RouterAgent) CompletionStream(prompt string, outputChan chan Response, processingChan chan interface{}, errChan chan error, ctx context.Context) {
	request := Request{
		Prompt:            prompt,
		Model:             agent.model,
		ResponseFormat:    agent.config.ResponseFormat,
		Stop:              agent.config.Stop,
		MaxTokens:         agent.config.MaxTokens,
		Temperature:       agent.config.Temperature,
		Tools:             agent.config.Tools,
		ToolChoice:        agent.config.ToolChoice,
		Seed:              agent.config.Seed,
		TopP:              agent.config.TopP,
		TopK:              agent.config.TopK,
		FrequencyPenalty:  agent.config.FrequencyPenalty,
		PresencePenalty:   agent.config.PresencePenalty,
		RepetitionPenalty: agent.config.RepetitionPenalty,
		LogitBias:         agent.config.LogitBias,
		TopLogprobs:       agent.config.TopLogprobs,
		MinP:              agent.config.MinP,
		TopA:              agent.config.TopA,
		Stream:            true,
	}

	agent.client.FetchChatCompletionsStream(request, outputChan, processingChan, errChan, ctx)
}

func (agent RouterAgent) Chat(messages []MessageRequest) (*Response, error) {
	request := Request{
		Messages:          messages,
		Model:             agent.model,
		ResponseFormat:    agent.config.ResponseFormat,
		Stop:              agent.config.Stop,
		MaxTokens:         agent.config.MaxTokens,
		Temperature:       agent.config.Temperature,
		Tools:             agent.config.Tools,
		ToolChoice:        agent.config.ToolChoice,
		Seed:              agent.config.Seed,
		TopP:              agent.config.TopP,
		TopK:              agent.config.TopK,
		FrequencyPenalty:  agent.config.FrequencyPenalty,
		PresencePenalty:   agent.config.PresencePenalty,
		RepetitionPenalty: agent.config.RepetitionPenalty,
		LogitBias:         agent.config.LogitBias,
		TopLogprobs:       agent.config.TopLogprobs,
		MinP:              agent.config.MinP,
		TopA:              agent.config.TopA,
		Stream:            false,
	}

	return agent.client.FetchChatCompletions(request)
}

func (agent RouterAgent) ChatStream(messages []MessageRequest, outputChan chan Response, processingChan chan interface{}, errChan chan error, ctx context.Context) {
	request := Request{
		Messages:          messages,
		Model:             agent.model,
		ResponseFormat:    agent.config.ResponseFormat,
		Stop:              agent.config.Stop,
		MaxTokens:         agent.config.MaxTokens,
		Temperature:       agent.config.Temperature,
		Tools:             agent.config.Tools,
		ToolChoice:        agent.config.ToolChoice,
		Seed:              agent.config.Seed,
		TopP:              agent.config.TopP,
		TopK:              agent.config.TopK,
		FrequencyPenalty:  agent.config.FrequencyPenalty,
		PresencePenalty:   agent.config.PresencePenalty,
		RepetitionPenalty: agent.config.RepetitionPenalty,
		LogitBias:         agent.config.LogitBias,
		TopLogprobs:       agent.config.TopLogprobs,
		MinP:              agent.config.MinP,
		TopA:              agent.config.TopA,
		Stream:            true,
	}

	agent.client.FetchChatCompletionsStream(request, outputChan, processingChan, errChan, ctx)
}

type message interface {
	GetRole() MessageRole
	GetContentPart() []ContentPart
	GetToolCalls() []ToolCall
	GetReasoning() string
	GetToolCallId() string
	GetName() string
}

type RouterAgentChat struct {
	RouterAgent
	Messages     []message
	ToolRegistry ToolRegistry
}

func NewRouterAgentChat(client *OpenRouterClient, model string, config RouterAgentConfig, system_prompt string) RouterAgentChat {
	return RouterAgentChat{
		RouterAgent: RouterAgent{
			client: client,
			model:  model,
			config: config,
		},
		Messages: []message{
			MessageRequest{
				Role:    RoleSystem,
				Content: system_prompt,
			},
		},
		ToolRegistry: *NewToolRegistry(),
	}
}

func AddToolToAgent[T any](agent *RouterAgentChat, definition ToolDefinition[T]) {
	agent.ToolRegistry.Register(toolWrapper[T]{
		definition: definition,
	})
}

func (agent *RouterAgentChat) generateMessagesForRequest() []MessageRequest {
	messages := make([]MessageRequest, 0, len(agent.Messages))
	for _, msg := range agent.Messages {
		messages = append(messages, MessageRequest{
			Role:       msg.GetRole(),
			Content:    msg.GetContentPart(),
			ToolCallID: msg.GetToolCallId(),
		})
	}
	return messages
}

func (agent *RouterAgentChat) callTools(response *Response) {
	for _, tool := range response.Choices[0].Message.ToolCalls {
		toolOutput, err := agent.ToolRegistry.CallTool(tool.Function.Name, json.RawMessage(tool.Function.Arguments))
		type errorOutput struct {
			Err string `json:"error"`
		}
		if err != nil {
			toolOutputByte, _ := json.Marshal(errorOutput{
				Err: fmt.Sprintf("%s", err),
			})
			toolOutput = string(toolOutputByte)
		}
		agent.Messages = append(agent.Messages, MessageRequest{
			Role: RoleTool,
			Content: []ContentPart{
				{
					Type: ContentTypeText,
					Text: toolOutput,
				},
			},
			ToolCallID: tool.ID,
			Name:       tool.Function.Name,
		})
	}
}

func (agent *RouterAgentChat) Chat(message string) error {
	agent.Messages = append(agent.Messages, MessageRequest{
		Role:    RoleUser,
		Content: message,
	})
	tools, err := agent.ToolRegistry.GenerateTools()
	if err != nil {
		return fmt.Errorf("error while generating tools: %s", err)
	}

	request := Request{
		Messages:          agent.generateMessagesForRequest(),
		Model:             agent.model,
		ResponseFormat:    agent.config.ResponseFormat,
		Stop:              agent.config.Stop,
		MaxTokens:         agent.config.MaxTokens,
		Temperature:       agent.config.Temperature,
		Tools:             tools,
		ToolChoice:        agent.config.ToolChoice,
		Seed:              agent.config.Seed,
		TopP:              agent.config.TopP,
		TopK:              agent.config.TopK,
		FrequencyPenalty:  agent.config.FrequencyPenalty,
		PresencePenalty:   agent.config.PresencePenalty,
		RepetitionPenalty: agent.config.RepetitionPenalty,
		LogitBias:         agent.config.LogitBias,
		TopLogprobs:       agent.config.TopLogprobs,
		MinP:              agent.config.MinP,
		TopA:              agent.config.TopA,
		Stream:            false,
	}

	response, err := agent.client.FetchChatCompletions(request)

	if err != nil {
		// rollback user message
		agent.Messages = agent.Messages[:len(agent.Messages)-1]
		return err
	}

	agent.Messages = append(agent.Messages, response.Choices[0].Message)

	agent.callTools(response)

	return nil
}

// https://openrouter.ai/docs/features/images-and-pdfs
func (agent *RouterAgentChat) ChatWithImage(message string, imgs ...image.Image) error {
	contentList := make([]ContentPart, 0, len(imgs)+1)
	contentList = append(contentList, ContentPart{
		Type: ContentTypeText,
		Text: message,
	})
	for _, img := range imgs {
		encodedImage, err := encodeImageToBase64(img)
		contentList = append(contentList, ContentPart{
			Type: ContentTypeImage,
			ImageURL: &ImageURL{
				URL: fmt.Sprintf("data:image/jpeg;base64,%s", encodedImage),
			},
		})
		if err != nil {
			return err
		}
	}

	agent.Messages = append(
		agent.Messages,
		MessageRequest{
			Role:    RoleUser,
			Content: contentList,
		})

	tools, err := agent.ToolRegistry.GenerateTools()
	if err != nil {
		return fmt.Errorf("error while generating tools: %s", err)
	}

	request := Request{
		Messages:          agent.generateMessagesForRequest(),
		Model:             agent.model,
		ResponseFormat:    agent.config.ResponseFormat,
		Stop:              agent.config.Stop,
		MaxTokens:         agent.config.MaxTokens,
		Temperature:       agent.config.Temperature,
		Tools:             tools,
		ToolChoice:        agent.config.ToolChoice,
		Seed:              agent.config.Seed,
		TopP:              agent.config.TopP,
		TopK:              agent.config.TopK,
		FrequencyPenalty:  agent.config.FrequencyPenalty,
		PresencePenalty:   agent.config.PresencePenalty,
		RepetitionPenalty: agent.config.RepetitionPenalty,
		LogitBias:         agent.config.LogitBias,
		TopLogprobs:       agent.config.TopLogprobs,
		MinP:              agent.config.MinP,
		TopA:              agent.config.TopA,
		Stream:            false,
	}

	response, err := agent.client.FetchChatCompletions(request)

	if err != nil {
		// rollback user message
		agent.Messages = agent.Messages[:len(agent.Messages)-1]
		return err
	}

	agent.Messages = append(agent.Messages, MessageRequest{
		Role:    RoleAssistant,
		Content: response.Choices[0].Message.Content,
	})

	return nil
}

func (agent *RouterAgentChat) ChatWithPDF(message string, pathsToPdf ...string) error {
	contentList := make([]ContentPart, 0, len(pathsToPdf)+1)
	contentList = append(contentList, ContentPart{
		Type: ContentTypeText,
		Text: message,
	})
	for _, pdf_path := range pathsToPdf {
		encodedPdf, err := encodePDFToBase64(pdf_path)
		contentList = append(contentList, ContentPart{
			Type: ContentTypePDF,
			File: &FileURL{
				Filename: pdf_path,
				FileData: fmt.Sprintf("data:application/pdf;base64,%s", encodedPdf),
			},
		})
		if err != nil {
			return err
		}
	}

	agent.Messages = append(
		agent.Messages,
		MessageRequest{
			Role:    RoleUser,
			Content: contentList,
		})

	tools, err := agent.ToolRegistry.GenerateTools()
	if err != nil {
		return fmt.Errorf("error while generating tools: %s", err)
	}

	request := Request{
		Messages:          agent.generateMessagesForRequest(),
		Model:             agent.model,
		ResponseFormat:    agent.config.ResponseFormat,
		Stop:              agent.config.Stop,
		MaxTokens:         agent.config.MaxTokens,
		Temperature:       agent.config.Temperature,
		Tools:             tools,
		ToolChoice:        agent.config.ToolChoice,
		Seed:              agent.config.Seed,
		TopP:              agent.config.TopP,
		TopK:              agent.config.TopK,
		FrequencyPenalty:  agent.config.FrequencyPenalty,
		PresencePenalty:   agent.config.PresencePenalty,
		RepetitionPenalty: agent.config.RepetitionPenalty,
		LogitBias:         agent.config.LogitBias,
		TopLogprobs:       agent.config.TopLogprobs,
		MinP:              agent.config.MinP,
		TopA:              agent.config.TopA,
		Stream:            false,
	}

	response, err := agent.client.FetchChatCompletions(request)

	if err != nil {
		// rollback user message
		agent.Messages = agent.Messages[:len(agent.Messages)-1]
		return err
	}

	agent.Messages = append(agent.Messages, MessageRequest{
		Role:    RoleAssistant,
		Content: response.Choices[0].Message.Content,
	})

	return nil
}
