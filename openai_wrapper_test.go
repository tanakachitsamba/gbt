package main

import (
	"context"
	"errors"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

type capturingStream struct{}

func (c *capturingStream) Recv() (openai.ChatCompletionStreamResponse, error) {
	return openai.ChatCompletionStreamResponse{}, nil
}

func (c *capturingStream) Close() {}

func TestOpenAIClient_CreateChatCompletionStream(t *testing.T) {
	expected := openai.ChatCompletionRequest{Model: "test"}
	dummyStream := &openai.ChatCompletionStream{}
	var wrapped bool

	client := newOpenAIClientWithHooks(
		func(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error) {
			require.Equal(t, expected, request)
			return dummyStream, nil
		},
		func(stream *openai.ChatCompletionStream) ChatCompletionStream {
			require.Equal(t, dummyStream, stream)
			wrapped = true
			return &capturingStream{}
		},
	)

	stream, err := client.CreateChatCompletionStream(context.Background(), expected)
	require.NoError(t, err)
	require.NotNil(t, stream)
	require.True(t, wrapped)
}

func TestOpenAIClient_CreateChatCompletionStreamError(t *testing.T) {
	expectedErr := errors.New("failed")
	client := newOpenAIClientWithHooks(
		func(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error) {
			return nil, expectedErr
		},
		func(stream *openai.ChatCompletionStream) ChatCompletionStream {
			t.Fatal("wrap should not be called when create fails")
			return nil
		},
	)

	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{})
	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, stream)
}

type fakeOpenAIAPI struct {
	called bool
}

func (f *fakeOpenAIAPI) CreateChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error) {
	f.called = true
	return &openai.ChatCompletionStream{}, nil
}

func TestNewOpenAIClient_UsesOpenAIStreamWrapper(t *testing.T) {
	api := &fakeOpenAIAPI{}

	wrapper := NewOpenAIClient(api)

	stream, err := wrapper.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{})
	require.NoError(t, err)
	require.True(t, api.called)
	require.IsType(t, &openAIStream{}, stream)
}
