package openai

import (
	"encoding/json"

	sdkopenai "github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared/constant"
)

type Model string

const (
	ModelGPT4o     Model = "gpt-4o"
	ModelGPT4oMini Model = "gpt-4o-mini"
	ModelGPT41Mini Model = "gpt-4.1-mini"
	ModelGPT41Nano Model = "gpt-4.1-nano"
)

func (m Model) String() string {
	return string(m)
}

func (m Model) orDefault() Model {
	if m == "" {
		return ModelGPT4oMini
	}
	return m
}

type ToolDefinition struct {
	param responses.ToolUnionParam
}

func (t ToolDefinition) toParam() responses.ToolUnionParam {
	return t.param
}

func RawToolDefinition(param responses.ToolUnionParam) ToolDefinition {
	return ToolDefinition{param: param}
}

type FunctionToolOption func(*responses.FunctionToolParam)

func WithFunctionDescription(description string) FunctionToolOption {
	return func(p *responses.FunctionToolParam) {
		if description != "" {
			p.Description = sdkopenai.String(description)
		}
	}
}

func WithFunctionStrict(strict bool) FunctionToolOption {
	return func(p *responses.FunctionToolParam) {
		p.Strict = param.NewOpt(strict)
	}
}

func NewFunctionTool(name string, parameters map[string]any, opts ...FunctionToolOption) ToolDefinition {
	tool := responses.FunctionToolParam{
		Name:       name,
		Parameters: parameters,
		Type:       constant.Function("function"),
		Strict:     param.NewOpt(true),
	}
	for _, opt := range opts {
		opt(&tool)
	}
	return ToolDefinition{param: responses.ToolUnionParam{OfFunction: &tool}}
}

type ResponseConfig struct {
	Model           Model
	Temperature     *float32
	Instructions    string
	MaxOutputTokens *int
	Tools           []ToolDefinition
}

type ResponseRequest struct {
	Input     string
	Config    ResponseConfig
	Callbacks StreamCallbacks
}

type StreamCallbacks struct {
	OnTextDelta     func(string)
	OnToolCallDelta func(ToolCallDelta)
}

type ResponseResult struct {
	Text      string
	ToolCalls []ToolCall
}

type ToolCall struct {
	ItemID        string
	CallID        string
	Name          string
	Arguments     string
	ArgumentsJSON json.RawMessage
}

type ToolCallDelta struct {
	ItemID         string
	CallID         string
	Name           string
	ArgumentsDelta string
	Arguments      string
	Completed      bool
}

func IntPtr(v int) *int { return &v }

func Float32Ptr(v float32) *float32 { return &v }
