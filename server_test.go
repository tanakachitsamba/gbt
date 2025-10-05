package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

func TestHandleRequest_Success(t *testing.T) {
	original := getClientFunc
	defer func() { getClientFunc = original }()

	stream := &fakeChatStream{
		responses: []openai.ChatCompletionStreamResponse{
			{
				Choices: []openai.ChatCompletionStreamChoice{{
					Delta:        openai.ChatCompletionStreamChoiceDelta{Content: "Answer"},
					FinishReason: "stop",
				}},
			},
		},
	}

	getClientFunc = func() ChatCompletionClient {
		return &fakeChatClient{stream: stream}
	}

	reqBody, err := json.Marshal(Request{Query: "Best soil"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handleRequest(rec, req)

	res := rec.Result()
	t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, "application/json", res.Header.Get("Content-Type"))
	require.Equal(t, "http://localhost:3000", res.Header.Get("Access-Control-Allow-Origin"))

	var payload Response
	err = json.NewDecoder(res.Body).Decode(&payload)
	require.NoError(t, err)
	require.Equal(t, "Answer", payload.Result)
	require.True(t, stream.closed)
}

func TestHandleRequest_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()

	handleRequest(rec, req)

	res := rec.Result()
	t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}
