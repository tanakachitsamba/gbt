package main

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// ChatCompletionClient represents the subset of the OpenAI client used by the application.
type ChatCompletionClient interface {
	CreateChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest) (ChatCompletionStream, error)
}

// ChatCompletionStream abstracts the streaming behaviour returned by the OpenAI client.
type ChatCompletionStream interface {
	Recv() (openai.ChatCompletionStreamResponse, error)
	Close()
}

// openAIClient is an adapter that wraps the official OpenAI client to satisfy ChatCompletionClient.
type openAIClient struct {
	createStream func(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error)
	wrapStream   func(*openai.ChatCompletionStream) ChatCompletionStream
}

// NewOpenAIClient creates a new ChatCompletionClient backed by the official OpenAI client.
func NewOpenAIClient(api interface {
	CreateChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error)
}) ChatCompletionClient {
	return &openAIClient{
		createStream: func(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error) {
			return api.CreateChatCompletionStream(ctx, request)
		},
		wrapStream: func(stream *openai.ChatCompletionStream) ChatCompletionStream {
			return &openAIStream{stream: stream}
		},
	}
}

// CreateChatCompletionStream delegates to the wrapped OpenAI client and adapts the result to ChatCompletionStream.
func (o *openAIClient) CreateChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest) (ChatCompletionStream, error) {
	stream, err := o.createStream(ctx, request)
	if err != nil {
		return nil, err
	}

	return o.wrapStream(stream), nil
}

// openAIStream is an adapter around openai.ChatCompletionStream to satisfy ChatCompletionStream.
type openAIStream struct {
	stream *openai.ChatCompletionStream
}

func (o *openAIStream) Recv() (openai.ChatCompletionStreamResponse, error) {
	return o.stream.Recv()
}

func (o *openAIStream) Close() {
	o.stream.Close()
}

func newOpenAIClientWithHooks(
	create func(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error),
	wrap func(*openai.ChatCompletionStream) ChatCompletionStream,
) *openAIClient {
	return &openAIClient{
		createStream: create,
		wrapStream:   wrap,
	}
}
