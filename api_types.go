package main

import "time"

// ResponseRequestV1 represents the payload for the v1 responses endpoint.
type ResponseRequestV1 struct {
	Model          string               `json:"model"`
	Instructions   string               `json:"instructions,omitempty"`
	Temperature    *float32             `json:"temperature,omitempty"`
	ResponseFormat *ResponseFormat      `json:"response_format,omitempty"`
	Tools          []ToolDefinition     `json:"tools,omitempty"`
	Files          []FileReference      `json:"files,omitempty"`
	ThreadID       string               `json:"thread_id,omitempty"`
	RunID          string               `json:"run_id,omitempty"`
	Input          []ResponseInputBlock `json:"input,omitempty"`
}

// ResponseInputBlock represents a block of input content provided by the caller.
type ResponseInputBlock struct {
	Role    string           `json:"role"`
	Content []MessageContent `json:"content"`
}

// ResponseFormat describes how the caller would like output to be formatted.
type ResponseFormat struct {
	Type string `json:"type"`
}

// ToolDefinition describes a callable tool exposed to the assistant.
type ToolDefinition struct {
	Type     string              `json:"type"`
	Function *FunctionDefinition `json:"function,omitempty"`
}

// FunctionDefinition describes a function tool signature.
type FunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// FileReference represents an uploaded file available to the assistant.
type FileReference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ResponseMessageV1 is the top-level response payload for v1 responses.
type ResponseMessageV1 struct {
	ID           string           `json:"id"`
	Object       string           `json:"object"`
	Created      int64            `json:"created"`
	Model        string           `json:"model"`
	Output       []ResponseOutput `json:"output"`
	Usage        *ResponseUsage   `json:"usage,omitempty"`
	ThreadID     string           `json:"thread_id,omitempty"`
	RunID        string           `json:"run_id,omitempty"`
	Instructions string           `json:"instructions,omitempty"`
}

// ResponseOutput represents the generated output blocks.
type ResponseOutput struct {
	ID      string           `json:"id"`
	Type    string           `json:"type"`
	Role    string           `json:"role,omitempty"`
	Status  string           `json:"status"`
	Content []MessageContent `json:"content"`
}

// MessageContent represents text-based content.
type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ResponseUsage mirrors the usage block returned by OpenAI.
type ResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ThreadRequestV1 represents the payload for creating a thread.
type ThreadRequestV1 struct {
	Title         string            `json:"title,omitempty"`
	Instructions  string            `json:"instructions,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	ToolResources map[string]any    `json:"tool_resources,omitempty"`
}

// ThreadResponseV1 is returned after creating a thread.
type ThreadResponseV1 struct {
	ID            string            `json:"id"`
	Object        string            `json:"object"`
	CreatedAt     time.Time         `json:"created_at"`
	Title         string            `json:"title,omitempty"`
	Instructions  string            `json:"instructions,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	ToolResources map[string]any    `json:"tool_resources,omitempty"`
	Status        string            `json:"status"`
}

// AssistantRequestV1 represents the payload for creating an assistant.
type AssistantRequestV1 struct {
	Name         string            `json:"name"`
	Model        string            `json:"model"`
	Instructions string            `json:"instructions,omitempty"`
	Tools        []ToolDefinition  `json:"tools,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// AssistantResponseV1 describes the assistant resource.
type AssistantResponseV1 struct {
	ID           string            `json:"id"`
	Object       string            `json:"object"`
	CreatedAt    time.Time         `json:"created_at"`
	Name         string            `json:"name"`
	Model        string            `json:"model"`
	Instructions string            `json:"instructions,omitempty"`
	Tools        []ToolDefinition  `json:"tools,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// VectorStoreRequestV1 represents the payload for creating a vector store.
type VectorStoreRequestV1 struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// VectorStoreResponseV1 represents the resulting vector store resource.
type VectorStoreResponseV1 struct {
	ID          string            `json:"id"`
	Object      string            `json:"object"`
	CreatedAt   time.Time         `json:"created_at"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Status      string            `json:"status"`
}
