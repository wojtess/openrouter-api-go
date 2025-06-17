package openrouterapigo

import "context"

type RouterAgentConfig struct {
	ResponseFormat    *ResponseFormat `json:"response_format,omitempty"`
	Stop              []string        `json:"stop,omitempty"`
	MaxTokens         int             `json:"max_tokens,omitempty"`
	Temperature       float64         `json:"temperature,omitempty"`
	Tools             []Tool          `json:"tools,omitempty"`
	ToolChoice        ToolChoice      `json:"tool_choice,omitempty"`
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

type RouterAgentChat struct {
	RouterAgent
	Messages []MessageRequest
}

func NewRouterAgentChat(client *OpenRouterClient, model string, config RouterAgentConfig, system_prompt string) RouterAgentChat {
	return RouterAgentChat{
		RouterAgent: RouterAgent{
			client: client,
			model:  model,
			config: config,
		},
		Messages: []MessageRequest{
			{
				Role:    RoleSystem,
				Content: system_prompt,
			},
		},
	}
}

func (agent *RouterAgentChat) Chat(message string) error {
	agent.Messages = append(agent.Messages, MessageRequest{
		Role:    RoleUser,
		Content: message,
	})
	request := Request{
		Messages:          agent.Messages,
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
