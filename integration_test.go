package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

type streamFixture struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
}

func TestAssistedWorkflowIntegration(t *testing.T) {
	original := getClientFunc
	defer func() { getClientFunc = original }()

	fixturePath := filepath.Join("testdata", "chat_stream_fixture.json")
	data, err := os.ReadFile(fixturePath)
	require.NoError(t, err)

	var chunks []streamFixture
	require.NoError(t, json.Unmarshal(data, &chunks))

	responses := make([]openai.ChatCompletionStreamResponse, len(chunks))
	for i, chunk := range chunks {
		responses[i] = openai.ChatCompletionStreamResponse{
			Choices: []openai.ChatCompletionStreamChoice{{
				Delta:        openai.ChatCompletionStreamChoiceDelta{Content: chunk.Content},
				FinishReason: chunk.FinishReason,
			}},
		}
	}

	stream := &fakeChatStream{responses: responses}
	getClientFunc = func() ChatCompletionClient {
		return &fakeChatClient{stream: stream}
	}

	server := httptest.NewServer(http.HandlerFunc(handleRequest))
	defer server.Close()

	resp, err := http.Post(server.URL, "application/json", strings.NewReader(`{"query":"Tomatoes"}`))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, resp.Body.Close()) })

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var payload Response
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.Equal(t, "First part second part", payload.Result)
	require.True(t, stream.closed)
}
