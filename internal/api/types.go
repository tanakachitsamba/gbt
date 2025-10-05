package api

import "errors"

// CreateResponseRequest captures the configurable parameters when invoking the
// Responses API via the HTTP surface.
type CreateResponseRequest struct {
	Model              string            `json:"model"`
	Messages           []Message         `json:"messages"`
	Instructions       string            `json:"instructions,omitempty"`
	Temperature        *float64          `json:"temperature,omitempty"`
	TopP               *float64          `json:"top_p,omitempty"`
	MaxOutputTokens    *int              `json:"max_output_tokens,omitempty"`
	MaxToolCalls       *int              `json:"max_tool_calls,omitempty"`
	Store              *bool             `json:"store,omitempty"`
	ParallelToolCalls  *bool             `json:"parallel_tool_calls,omitempty"`
	ToolChoice         *ToolChoice       `json:"tool_choice,omitempty"`
	Tools              []ToolDefinition  `json:"tools,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	Include            []string          `json:"include,omitempty"`
	ServiceTier        string            `json:"service_tier,omitempty"`
	PreviousResponseID string            `json:"previous_response_id,omitempty"`
	PromptCacheKey     string            `json:"prompt_cache_key,omitempty"`
	SafetyIdentifier   string            `json:"safety_identifier,omitempty"`
	User               string            `json:"user,omitempty"`
	Stream             bool              `json:"stream,omitempty"`
}

// Validate performs basic sanity checks on the request payload prior to
// translating it into the SDK specific types.
func (r *CreateResponseRequest) Validate() error {
	if r == nil {
		return errors.New("request body is required")
	}
	if r.Model == "" {
		return errors.New("model is required")
	}
	if len(r.Messages) == 0 {
		return errors.New("at least one message is required")
	}
	for i := range r.Messages {
		if err := r.Messages[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Message represents a conversational turn supplied to the Responses API.
type Message struct {
	Role    string           `json:"role"`
	Content []MessageContent `json:"content,omitempty"`
	Text    string           `json:"text,omitempty"`
}

// Validate ensures the message contains a supported role and content.
func (m *Message) Validate() error {
	if m == nil {
		return errors.New("message cannot be null")
	}
	switch m.Role {
	case "user", "system", "assistant", "developer":
	default:
		return errors.New("message role must be one of user, assistant, system, or developer")
	}

	if len(m.Content) == 0 && m.Text == "" {
		return errors.New("message content or text must be provided")
	}
	for i := range m.Content {
		if err := m.Content[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

// MessageContent represents a single content block within a message. Only text
// content is currently supported, but the structure allows for future expansion
// into multimodal payloads.
type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Validate performs minimal validation on an individual content block.
func (c *MessageContent) Validate() error {
	if c == nil {
		return errors.New("content item cannot be null")
	}
	switch c.Type {
	case "text", "input_text", "":
		if c.Text == "" {
			return errors.New("text content must include a non-empty text field")
		}
	default:
		return errors.New("unsupported content type: " + c.Type)
	}
	return nil
}

// ToolDefinition captures a custom tool/function definition that should be made
// available to the model.
type ToolDefinition struct {
	Type     string                 `json:"type"`
	Function *FunctionTool          `json:"function,omitempty"`
	Raw      map[string]any         `json:"raw,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FunctionTool describes a single function the model can call.
type FunctionTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Strict      *bool          `json:"strict,omitempty"`
}

// ToolChoice allows callers to control whether the model should call a tool and
// optionally force a specific function to be executed.
type ToolChoice struct {
	Type     string              `json:"type"`
	Function *ToolChoiceFunction `json:"function,omitempty"`
}

// ToolChoiceFunction specifies the name of the function to call when the tool
// choice type is set to "function".
type ToolChoiceFunction struct {
	Name string `json:"name"`
}

// ErrorResponse models the JSON response returned when a request fails.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains the user facing error details.
type ErrorBody struct {
	Message string `json:"message"`
}
