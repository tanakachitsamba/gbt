package main

import (
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
)

type fakeChatClient struct {
	stream       ChatCompletionStream
	err          error
	capturedCtx  context.Context
	capturedReq  openai.ChatCompletionRequest
	createCalled bool
}

func (f *fakeChatClient) CreateChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest) (ChatCompletionStream, error) {
	f.createCalled = true
	f.capturedCtx = ctx
	f.capturedReq = request
	if f.err != nil {
		return nil, f.err
	}
	if f.stream == nil {
		return nil, errors.New("no stream configured")
	}
	return f.stream, nil
}

type fakeChatStream struct {
	responses []openai.ChatCompletionStreamResponse
	err       error
	errIndex  int
	idx       int
	closed    bool
}

func (f *fakeChatStream) Recv() (openai.ChatCompletionStreamResponse, error) {
	if f.err != nil && f.idx == f.errIndex {
		return openai.ChatCompletionStreamResponse{}, f.err
	}
	if f.idx >= len(f.responses) {
		return openai.ChatCompletionStreamResponse{}, errors.New("out of responses")
	}
	resp := f.responses[f.idx]
	f.idx++
	return resp, nil
}

func (f *fakeChatStream) Close() {
	f.closed = true
}
