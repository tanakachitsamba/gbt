package openaiclient

import (
	"context"
	"errors"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/openai/openai-go/responses"
)

// Config contains the minimal configuration required to construct a new OpenAI
// client. Additional request options can be supplied when creating the client
// if required by downstream callers.
type Config struct {
	APIKey       string
	BaseURL      string
	Organization string
	Project      string
}

// Client wraps the generated OpenAI SDK client and exposes higher level helper
// methods that the HTTP handlers can rely on without depending directly on the
// generated types.
type Client struct {
	client *openai.Client
}

// New creates a new OpenAI client using the supplied configuration and optional
// request options. The configuration mirrors the standard environment variables
// supported by the official SDK.
func New(cfg Config, opts ...option.RequestOption) (*Client, error) {
	requestOpts := make([]option.RequestOption, 0, 4+len(opts))

	if cfg.APIKey != "" {
		requestOpts = append(requestOpts, option.WithAPIKey(cfg.APIKey))
	}
	if cfg.BaseURL != "" {
		requestOpts = append(requestOpts, option.WithBaseURL(cfg.BaseURL))
	}
	if cfg.Organization != "" {
		requestOpts = append(requestOpts, option.WithOrganization(cfg.Organization))
	}
	if cfg.Project != "" {
		requestOpts = append(requestOpts, option.WithProject(cfg.Project))
	}

	requestOpts = append(requestOpts, opts...)

	raw := openai.NewClient(requestOpts...)
	return &Client{client: &raw}, nil
}

// NewFromEnvironment creates a client using the standard OpenAI environment
// variables. Both OPENAI_API_KEY and the legacy OPENAI_KEY are supported to
// ease migration from older projects.
func NewFromEnvironment(opts ...option.RequestOption) (*Client, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_KEY")
	}
	if apiKey == "" {
		return nil, errors.New("missing OpenAI API key: set OPENAI_API_KEY or OPENAI_KEY")
	}

	cfg := Config{
		APIKey:       apiKey,
		BaseURL:      os.Getenv("OPENAI_BASE_URL"),
		Organization: os.Getenv("OPENAI_ORG_ID"),
		Project:      os.Getenv("OPENAI_PROJECT_ID"),
	}

	return New(cfg, opts...)
}

// Responses exposes a thin wrapper around the official SDK's Responses service
// so that handlers can work with a simplified interface that is easy to mock in
// unit tests.
func (c *Client) Responses() ResponsesClient {
	return &responsesClient{svc: &c.client.Responses}
}

// ResponsesClient defines the operations required by the HTTP layer for working
// with the Responses API.
type ResponsesClient interface {
	CreateResponse(ctx context.Context, params responses.ResponseNewParams) (*ResponseResult, error)
	StreamResponse(ctx context.Context, params responses.ResponseNewParams) (ResponseStream, error)
}

// ResponseResult provides a convenient representation of a created response,
// including the aggregated output text for callers that only care about the
// final assistant message.
type ResponseResult struct {
	Response   *responses.Response `json:"response"`
	OutputText string              `json:"output_text"`
}

type responsesClient struct {
	svc *responses.ResponseService
}

func (c *responsesClient) CreateResponse(ctx context.Context, params responses.ResponseNewParams) (*ResponseResult, error) {
	resp, err := c.svc.New(ctx, params)
	if err != nil {
		return nil, err
	}
	return &ResponseResult{Response: resp, OutputText: resp.OutputText()}, nil
}

func (c *responsesClient) StreamResponse(ctx context.Context, params responses.ResponseNewParams) (ResponseStream, error) {
	stream := c.svc.NewStreaming(ctx, params)
	if err := stream.Err(); err != nil {
		// Close the stream to avoid leaking the underlying HTTP connection when the
		// request fails before streaming begins.
		_ = stream.Close()
		return nil, err
	}
	return &responseStream{stream: stream}, nil
}

// ResponseStream adapts the streaming iterator returned by the official SDK so
// it can be easily mocked during testing.
type ResponseStream interface {
	Next() bool
	Current() responses.ResponseStreamEventUnion
	Err() error
	Close() error
}

type responseStream struct {
	stream *ssestream.Stream[responses.ResponseStreamEventUnion]
}

func (s *responseStream) Next() bool {
	return s.stream.Next()
}

func (s *responseStream) Current() responses.ResponseStreamEventUnion {
	return s.stream.Current()
}

func (s *responseStream) Err() error {
	return s.stream.Err()
}

func (s *responseStream) Close() error {
	return s.stream.Close()
}
