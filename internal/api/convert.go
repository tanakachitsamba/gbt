package api

import (
	"fmt"

	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
)

// BuildResponseParams converts the high level request payload into the
// parameters required by the official OpenAI Go SDK.
func BuildResponseParams(req *CreateResponseRequest) (responses.ResponseNewParams, error) {
	var params responses.ResponseNewParams
	if req == nil {
		return params, fmt.Errorf("request cannot be nil")
	}

	params.Model = responses.ResponsesModel(req.Model)

	inputItems, err := buildInputItems(req.Messages)
	if err != nil {
		return params, err
	}
	params.Input = responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems}

	if req.Instructions != "" {
		params.Instructions = param.NewOpt(req.Instructions)
	}
	if req.Temperature != nil {
		params.Temperature = param.NewOpt(*req.Temperature)
	}
	if req.TopP != nil {
		params.TopP = param.NewOpt(*req.TopP)
	}
	if req.MaxOutputTokens != nil {
		params.MaxOutputTokens = param.NewOpt(int64(*req.MaxOutputTokens))
	}
	if req.MaxToolCalls != nil {
		params.MaxToolCalls = param.NewOpt(int64(*req.MaxToolCalls))
	}
	if req.Store != nil {
		params.Store = param.NewOpt(*req.Store)
	}
	if req.ParallelToolCalls != nil {
		params.ParallelToolCalls = param.NewOpt(*req.ParallelToolCalls)
	}
	if req.Metadata != nil {
		params.Metadata = shared.Metadata(req.Metadata)
	}
	if req.Include != nil {
		params.Include = make([]responses.ResponseIncludable, 0, len(req.Include))
		for _, include := range req.Include {
			params.Include = append(params.Include, responses.ResponseIncludable(include))
		}
	}
	if req.ServiceTier != "" {
		tier, err := parseServiceTier(req.ServiceTier)
		if err != nil {
			return params, err
		}
		params.ServiceTier = tier
	}
	if req.PreviousResponseID != "" {
		params.PreviousResponseID = param.NewOpt(req.PreviousResponseID)
	}
	if req.PromptCacheKey != "" {
		params.PromptCacheKey = param.NewOpt(req.PromptCacheKey)
	}
	if req.SafetyIdentifier != "" {
		params.SafetyIdentifier = param.NewOpt(req.SafetyIdentifier)
	}
	if req.User != "" {
		params.User = param.NewOpt(req.User)
	}

	if len(req.Tools) > 0 {
		tools, err := buildToolParams(req.Tools)
		if err != nil {
			return params, err
		}
		params.Tools = tools
	}
	if req.ToolChoice != nil {
		choice, err := buildToolChoice(*req.ToolChoice)
		if err != nil {
			return params, err
		}
		params.ToolChoice = choice
	}

	return params, nil
}

func buildInputItems(messages []Message) (responses.ResponseInputParam, error) {
	items := make(responses.ResponseInputParam, 0, len(messages))
	for _, msg := range messages {
		messageParam, err := buildMessageParam(msg)
		if err != nil {
			return nil, err
		}
		items = append(items, responses.ResponseInputItemUnionParam{OfMessage: messageParam})
	}
	return items, nil
}

func buildMessageParam(msg Message) (*responses.EasyInputMessageParam, error) {
	paramMsg := responses.EasyInputMessageParam{
		Role: responses.EasyInputMessageRole(msg.Role),
	}

	if len(msg.Content) == 0 {
		paramMsg.Content = responses.EasyInputMessageContentUnionParam{
			OfString: param.NewOpt(msg.Text),
		}
		return &paramMsg, nil
	}

	contentList, err := buildContentList(msg.Content)
	if err != nil {
		return nil, err
	}
	paramMsg.Content = responses.EasyInputMessageContentUnionParam{OfInputItemContentList: contentList}
	return &paramMsg, nil
}

func buildContentList(contents []MessageContent) (responses.ResponseInputMessageContentListParam, error) {
	list := make(responses.ResponseInputMessageContentListParam, 0, len(contents))
	for _, c := range contents {
		switch c.Type {
		case "", "text", "input_text":
			if c.Text == "" {
				return nil, fmt.Errorf("text content must not be empty")
			}
			list = append(list, responses.ResponseInputContentUnionParam{
				OfInputText: &responses.ResponseInputTextParam{Text: c.Text},
			})
		default:
			return nil, fmt.Errorf("unsupported content type: %s", c.Type)
		}
	}
	return list, nil
}

func buildToolParams(tools []ToolDefinition) ([]responses.ToolUnionParam, error) {
	out := make([]responses.ToolUnionParam, 0, len(tools))
	for _, tool := range tools {
		switch tool.Type {
		case "function":
			if tool.Function == nil {
				return nil, fmt.Errorf("function tool must include a function definition")
			}
			fn := responses.FunctionToolParam{
				Name:       tool.Function.Name,
				Parameters: map[string]any{},
			}
			if fn.Name == "" {
				return nil, fmt.Errorf("function tool name is required")
			}
			if tool.Function.Parameters != nil {
				fn.Parameters = tool.Function.Parameters
			}
			if tool.Function.Description != "" {
				fn.Description = param.NewOpt(tool.Function.Description)
			}
			strict := false
			if tool.Function.Strict != nil {
				strict = *tool.Function.Strict
			}
			fn.Strict = param.NewOpt(strict)
			out = append(out, responses.ToolUnionParam{OfFunction: &fn})
		default:
			return nil, fmt.Errorf("unsupported tool type: %s", tool.Type)
		}
	}
	return out, nil
}

func buildToolChoice(choice ToolChoice) (responses.ResponseNewParamsToolChoiceUnion, error) {
	var union responses.ResponseNewParamsToolChoiceUnion
	switch choice.Type {
	case "", "auto":
		union.OfToolChoiceMode = param.NewOpt(responses.ToolChoiceOptionsAuto)
	case "none":
		union.OfToolChoiceMode = param.NewOpt(responses.ToolChoiceOptionsNone)
	case "required":
		union.OfToolChoiceMode = param.NewOpt(responses.ToolChoiceOptionsRequired)
	case "function":
		if choice.Function == nil || choice.Function.Name == "" {
			return union, fmt.Errorf("tool_choice.function.name is required when forcing a function call")
		}
		union.OfFunctionTool = &responses.ToolChoiceFunctionParam{Name: choice.Function.Name}
	default:
		return union, fmt.Errorf("unsupported tool_choice type: %s", choice.Type)
	}
	return union, nil
}

func parseServiceTier(tier string) (responses.ResponseNewParamsServiceTier, error) {
	switch tier {
	case "auto":
		return responses.ResponseNewParamsServiceTierAuto, nil
	case "default":
		return responses.ResponseNewParamsServiceTierDefault, nil
	case "flex":
		return responses.ResponseNewParamsServiceTierFlex, nil
	case "scale":
		return responses.ResponseNewParamsServiceTierScale, nil
	case "priority":
		return responses.ResponseNewParamsServiceTierPriority, nil
	default:
		return "", fmt.Errorf("unsupported service_tier: %s", tier)
	}
}
