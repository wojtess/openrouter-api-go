package openrouterapigo

// Request represents the main request structure.
type Request struct {
	Messages          []MessageRequest     `json:"messages,omitempty"`
	Prompt            string               `json:"prompt,omitempty"`
	Model             string               `json:"model,omitempty"`
	ResponseFormat    *ResponseFormat      `json:"response_format,omitempty"`
	Stop              []string             `json:"stop,omitempty"`
	Stream            bool                 `json:"stream,omitempty"`
	MaxTokens         int                  `json:"max_tokens,omitempty"`
	Temperature       float64              `json:"temperature,omitempty"`
	Tools             []Tool               `json:"tools,omitempty"`
	ToolChoice        ToolChoice           `json:"tool_choice,omitempty"`
	Seed              int                  `json:"seed,omitempty"`
	TopP              float64              `json:"top_p,omitempty"`
	TopK              int                  `json:"top_k,omitempty"`
	FrequencyPenalty  float64              `json:"frequency_penalty,omitempty"`
	PresencePenalty   float64              `json:"presence_penalty,omitempty"`
	RepetitionPenalty float64              `json:"repetition_penalty,omitempty"`
	LogitBias         map[int]float64      `json:"logit_bias,omitempty"`
	TopLogprobs       int                  `json:"top_logprobs,omitempty"`
	MinP              float64              `json:"min_p,omitempty"`
	TopA              float64              `json:"top_a,omitempty"`
	Prediction        *Prediction          `json:"prediction,omitempty"`
	Transforms        []string             `json:"transforms,omitempty"`
	Models            []string             `json:"models,omitempty"`
	Route             string               `json:"route,omitempty"`
	Provider          *ProviderPreferences `json:"provider,omitempty"`
}

// ResponseFormat represents the response format structure.
type ResponseFormat struct {
	Type string `json:"type"`
}

// Prediction represents the prediction structure.
type Prediction struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// ProviderPreferences represents the provider preferences structure.
type ProviderPreferences struct {
	RefererURL string `json:"referer_url,omitempty"`
	SiteName   string `json:"site_name,omitempty"`
}

// Message represents the message structure.
type MessageRequest struct {
	Role       MessageRole `json:"role"`
	Content    interface{} `json:"content"` // Can be string or []ContentPart
	Name       string      `json:"name,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
}

type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// ContentPart represents the content part structure.
type ContentPart struct {
	Type     ContnetType `json:"type"`
	Text     string      `json:"text,omitempty"`
	ImageURL *ImageURL   `json:"image_url,omitempty"`
}

type ContnetType string

const (
	ContentTypeText  ContnetType = "text"
	ContentTypeImage ContnetType = "image_url"
)

// ImageURL represents the image URL structure.
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// FunctionDescription represents the function description structure.
type FunctionDescription struct {
	Description string      `json:"description,omitempty"`
	Name        string      `json:"name"`
	Parameters  interface{} `json:"parameters"` // JSON Schema object
}

// Tool represents the tool structure.
type Tool struct {
	Type     string              `json:"type"`
	Function FunctionDescription `json:"function"`
}

// ToolChoice represents the tool choice structure.
type ToolChoice struct {
	Type     string `json:"type"`
	Function struct {
		Name string `json:"name"`
	} `json:"function"`
}

type Response struct {
	ID                string         `json:"id"`
	Choices           []Choice       `json:"choices"`
	Created           int64          `json:"created"`
	Model             string         `json:"model"`
	Object            string         `json:"object"`
	SystemFingerprint *string        `json:"system_fingerprint,omitempty"`
	Usage             *ResponseUsage `json:"usage,omitempty"`
}

type ResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	FinishReason string           `json:"finish_reason"`
	Text         string           `json:"text,omitempty"`
	Message      *MessageResponse `json:"message,omitempty"`
	Delta        *Delta           `json:"delta,omitempty"`
	Error        *ErrorResponse   `json:"error,omitempty"`
}

type MessageResponse struct {
	Content   string     `json:"content"`
	Role      string     `json:"role"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type Delta struct {
	Content   string     `json:"content"`
	Role      string     `json:"role,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ErrorResponse struct {
	Code     int                    `json:"code"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ToolCall struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Function interface{} `json:"function"`
}
