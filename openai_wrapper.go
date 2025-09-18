package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

var (
	errMissingModel    = errors.New("model is required")
	errMissingInput    = errors.New("instructions or input content is required")
	errAssistantName   = errors.New("assistant name is required")
	errVectorStoreName = errors.New("vector store name is required")
)

// OpenAIWrapper centralises interactions with the OpenAI client so handlers can remain thin.
type OpenAIWrapper struct {
	client *openai.Client
}

// NewOpenAIWrapper constructs a new wrapper using the provided OpenAI client.
func NewOpenAIWrapper(client *openai.Client) *OpenAIWrapper {
	return &OpenAIWrapper{client: client}
}

// CreateResponse issues a chat completion request and maps the result to the v1 response DTO.
func (o *OpenAIWrapper) CreateResponse(ctx context.Context, req ResponseRequestV1) (ResponseMessageV1, error) {
	if req.Model == "" {
		return ResponseMessageV1{}, errMissingModel
	}

	messages := buildMessages(req)
	if len(messages) == 0 {
		return ResponseMessageV1{}, errMissingInput
	}

	temperature := float32(0.7)
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	completionReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: temperature,
	}

	completion, err := o.client.CreateChatCompletion(ctx, completionReq)
	if err != nil {
		return ResponseMessageV1{}, fmt.Errorf("create chat completion: %w", err)
	}

	output := make([]ResponseOutput, 0, len(completion.Choices))
	for _, choice := range completion.Choices {
		content := []MessageContent{
			{
				Type: "output_text",
				Text: choice.Message.Content,
			},
		}

		output = append(output, ResponseOutput{
			ID:      fmt.Sprintf("msg_%s", uuid.New().String()),
			Type:    "message",
			Role:    choice.Message.Role,
			Status:  "completed",
			Content: content,
		})
	}

	usage := &ResponseUsage{
		PromptTokens:     completion.Usage.PromptTokens,
		CompletionTokens: completion.Usage.CompletionTokens,
		TotalTokens:      completion.Usage.TotalTokens,
	}

	response := ResponseMessageV1{
		ID:           completion.ID,
		Object:       "response",
		Created:      completion.Created,
		Model:        completion.Model,
		Output:       output,
		Usage:        usage,
		ThreadID:     req.ThreadID,
		RunID:        req.RunID,
		Instructions: req.Instructions,
	}

	return response, nil
}

// CreateThread materialises a thread resource locally so the front-end can coordinate runs.
func (o *OpenAIWrapper) CreateThread(_ context.Context, req ThreadRequestV1) (ThreadResponseV1, error) {
	response := ThreadResponseV1{
		ID:            fmt.Sprintf("thread_%s", uuid.New().String()),
		Object:        "thread",
		CreatedAt:     time.Now().UTC(),
		Title:         req.Title,
		Instructions:  req.Instructions,
		Metadata:      req.Metadata,
		ToolResources: req.ToolResources,
		Status:        "open",
	}

	return response, nil
}

// CreateAssistant materialises an assistant resource locally.
func (o *OpenAIWrapper) CreateAssistant(_ context.Context, req AssistantRequestV1) (AssistantResponseV1, error) {
	if req.Name == "" {
		return AssistantResponseV1{}, errAssistantName
	}
	if req.Model == "" {
		return AssistantResponseV1{}, errMissingModel
	}

	response := AssistantResponseV1{
		ID:           fmt.Sprintf("asst_%s", uuid.New().String()),
		Object:       "assistant",
		CreatedAt:    time.Now().UTC(),
		Name:         req.Name,
		Model:        req.Model,
		Instructions: req.Instructions,
		Tools:        req.Tools,
		Metadata:     req.Metadata,
	}

	return response, nil
}

// CreateVectorStore materialises a vector store resource locally.
func (o *OpenAIWrapper) CreateVectorStore(_ context.Context, req VectorStoreRequestV1) (VectorStoreResponseV1, error) {
	if req.Name == "" {
		return VectorStoreResponseV1{}, errVectorStoreName
	}

	response := VectorStoreResponseV1{
		ID:          fmt.Sprintf("vs_%s", uuid.New().String()),
		Object:      "vector_store",
		CreatedAt:   time.Now().UTC(),
		Name:        req.Name,
		Description: req.Description,
		Metadata:    req.Metadata,
		Status:      "ready",
	}

	return response, nil
}

func buildMessages(req ResponseRequestV1) []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, 0)

	if req.Instructions != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.Instructions,
		})
	}

	for _, block := range req.Input {
		if block.Role == "" {
			continue
		}
		text := concatenateContent(block.Content)
		if text == "" {
			continue
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    block.Role,
			Content: text,
		})
	}

	hasUser := false
	for _, msg := range messages {
		if msg.Role == openai.ChatMessageRoleUser {
			hasUser = true
			break
		}
	}

	if !hasUser && req.Instructions != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: req.Instructions,
		})
	}

	return messages
}

func concatenateContent(content []MessageContent) string {
	if len(content) == 0 {
		return ""
	}

	var result string
	for _, block := range content {
		if block.Text == "" {
			continue
		}
		result += block.Text
	}

	return result
}
