package openai

import (
	"context"

	sdkopenai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
)

type Client struct {
	sdk *sdkopenai.Client
}

func NewClient(opts ...option.RequestOption) *Client {
	client := sdkopenai.NewClient(opts...)
	return &Client{sdk: &client}
}

func (c *Client) SDK() *sdkopenai.Client {
	if c == nil {
		return nil
	}
	return c.sdk
}

func (c *Client) Responses() *responses.ResponseService {
	if c == nil || c.sdk == nil {
		return nil
	}
	return &c.sdk.Responses
}

func (c *Client) Assistants() *sdkopenai.BetaAssistantService {
	if c == nil || c.sdk == nil {
		return nil
	}
	return &c.sdk.Beta.Assistants
}

func (c *Client) VectorStores() *sdkopenai.VectorStoreService {
	if c == nil || c.sdk == nil {
		return nil
	}
	return &c.sdk.VectorStores
}

func (c *Client) Uploads() *sdkopenai.UploadService {
	if c == nil || c.sdk == nil {
		return nil
	}
	return &c.sdk.Uploads
}

func (c *Client) Audio() *sdkopenai.AudioService {
	if c == nil || c.sdk == nil {
		return nil
	}
	return &c.sdk.Audio
}

func (c *Client) StreamResponse(ctx context.Context, req ResponseRequest) (ResponseResult, error) {
	params := buildResponseParams(req)
	svc := c.Responses()
	if svc == nil {
		return ResponseResult{}, ErrMissingClient
	}

	stream := svc.NewStreaming(ctx, params)
	if stream == nil {
		return ResponseResult{}, ErrStreamUnavailable
	}
	defer stream.Close()

	if err := stream.Err(); err != nil {
		return ResponseResult{}, err
	}

	return consumeResponseStream(stream, req.Callbacks)
}
