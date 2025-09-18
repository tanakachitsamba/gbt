package main

import (
	"errors"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

func TestGetChatStreamResponse_Success(t *testing.T) {
	stream := &fakeChatStream{
		responses: []openai.ChatCompletionStreamResponse{
			{
				Choices: []openai.ChatCompletionStreamChoice{{
					Delta: openai.ChatCompletionStreamChoiceDelta{Content: "Hello "},
				}},
			},
			{
				Choices: []openai.ChatCompletionStreamChoice{{
					Delta:        openai.ChatCompletionStreamChoiceDelta{Content: "garden"},
					FinishReason: "stop",
				}},
			},
		},
	}

	client := &fakeChatClient{stream: stream}
	inp := Input{
		client:        client,
		prompt:        "How to plant tulips?",
		model:         "test-model",
		temperature:   0.2,
		maxTokens:     42,
		systemMessage: "system",
	}

	result, err := inp.getChatStreamResponse()
	require.NoError(t, err)
	require.Equal(t, "Hello garden", result)
	require.True(t, stream.closed)
	require.True(t, client.createCalled)
	require.Equal(t, inp.maxTokens, client.capturedReq.MaxTokens)
	require.Equal(t, inp.temperature, client.capturedReq.Temperature)
	require.Equal(t, inp.prompt, client.capturedReq.Messages[1].Content)
}

func TestGetChatStreamResponse_CreateStreamError(t *testing.T) {
	client := &fakeChatClient{err: errors.New("boom")}
	inp := Input{client: client}

	_, err := inp.getChatStreamResponse()
	require.Error(t, err)
	require.True(t, client.createCalled)
}

func TestGetChatStreamResponse_StreamRecvError(t *testing.T) {
	stream := &fakeChatStream{
		responses: []openai.ChatCompletionStreamResponse{
			{
				Choices: []openai.ChatCompletionStreamChoice{{
					Delta: openai.ChatCompletionStreamChoiceDelta{Content: "partial"},
				}},
			},
		},
		err:      errors.New("recv failure"),
		errIndex: 1,
	}

	client := &fakeChatClient{stream: stream}
	inp := Input{client: client}

	_, err := inp.getChatStreamResponse()
	require.Error(t, err)
	require.True(t, stream.closed)
}
