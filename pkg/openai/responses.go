package openai

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	sdkopenai "github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/openai/openai-go/responses"
)

var (
	ErrMissingClient     = errors.New("openai client is not configured")
	ErrStreamUnavailable = errors.New("response stream unavailable")
)

func buildResponseParams(req ResponseRequest) responses.ResponseNewParams {
	cfg := req.Config
	params := responses.ResponseNewParams{
		Model: responses.ResponsesModel(cfg.Model.orDefault()),
	}

	if req.Input != "" {
		params.Input = responses.ResponseNewParamsInputUnion{
			OfString: sdkopenai.String(req.Input),
		}
	}

	if cfg.Instructions != "" {
		params.Instructions = sdkopenai.String(cfg.Instructions)
	}

	if cfg.Temperature != nil {
		params.Temperature = sdkopenai.Float(float64(*cfg.Temperature))
	}

	if cfg.MaxOutputTokens != nil {
		params.MaxOutputTokens = sdkopenai.Int(int64(*cfg.MaxOutputTokens))
	}

	if len(cfg.Tools) > 0 {
		params.Tools = make([]responses.ToolUnionParam, 0, len(cfg.Tools))
		for _, tool := range cfg.Tools {
			params.Tools = append(params.Tools, tool.toParam())
		}
	}

	return params
}

func consumeResponseStream(stream *ssestream.Stream[responses.ResponseStreamEventUnion], callbacks StreamCallbacks) (ResponseResult, error) {
	var textBuilder strings.Builder
	builders := map[string]*toolCallBuilder{}
	order := make([]string, 0)

	for stream.Next() {
		event := stream.Current()

		switch event.Type {
		case "response.output_text.delta":
			delta := event.AsResponseOutputTextDelta()
			textBuilder.WriteString(delta.Delta)
			if callbacks.OnTextDelta != nil {
				callbacks.OnTextDelta(delta.Delta)
			}
		case "response.output_text.done":
			done := event.AsResponseOutputTextDone()
			if textBuilder.Len() == 0 && done.Text != "" {
				textBuilder.WriteString(done.Text)
				if callbacks.OnTextDelta != nil {
					callbacks.OnTextDelta(done.Text)
				}
			}
		case "response.output_item.added":
			added := event.AsResponseOutputItemAdded()
			if functionCall, ok := added.Item.AsAny().(responses.ResponseFunctionToolCall); ok {
				builder, seen := builders[added.Item.ID]
				if !seen {
					builder = &toolCallBuilder{ItemID: added.Item.ID}
					builders[added.Item.ID] = builder
					order = append(order, added.Item.ID)
				}
				if functionCall.CallID != "" {
					builder.CallID = functionCall.CallID
				} else if functionCall.ID != "" {
					builder.CallID = functionCall.ID
				}
				if functionCall.Name != "" {
					builder.Name = functionCall.Name
				}
				if functionCall.Arguments != "" {
					if builder.Arguments.Len() == 0 {
						builder.Arguments.WriteString(functionCall.Arguments)
					}
					if callbacks.OnToolCallDelta != nil {
						callbacks.OnToolCallDelta(ToolCallDelta{
							ItemID:         builder.ItemID,
							CallID:         builder.CallID,
							Name:           builder.Name,
							ArgumentsDelta: functionCall.Arguments,
							Arguments:      builder.Arguments.String(),
							Completed:      builder.Completed,
						})
					}
				}
			}
		case "response.function_call_arguments.delta":
			delta := event.AsResponseFunctionCallArgumentsDelta()
			builder, seen := builders[delta.ItemID]
			if !seen {
				builder = &toolCallBuilder{ItemID: delta.ItemID}
				builders[delta.ItemID] = builder
				order = append(order, delta.ItemID)
			}
			builder.Arguments.WriteString(delta.Delta)
			if callbacks.OnToolCallDelta != nil {
				callbacks.OnToolCallDelta(ToolCallDelta{
					ItemID:         builder.ItemID,
					CallID:         builder.CallID,
					Name:           builder.Name,
					ArgumentsDelta: delta.Delta,
					Arguments:      builder.Arguments.String(),
					Completed:      builder.Completed,
				})
			}
		case "response.function_call_arguments.done":
			done := event.AsResponseFunctionCallArgumentsDone()
			builder, seen := builders[done.ItemID]
			if !seen {
				builder = &toolCallBuilder{ItemID: done.ItemID}
				builders[done.ItemID] = builder
				order = append(order, done.ItemID)
			}
			if done.Arguments != "" {
				builder.Arguments.Reset()
				builder.Arguments.WriteString(done.Arguments)
			}
			builder.Completed = true
			if callbacks.OnToolCallDelta != nil {
				callbacks.OnToolCallDelta(ToolCallDelta{
					ItemID:         builder.ItemID,
					CallID:         builder.CallID,
					Name:           builder.Name,
					ArgumentsDelta: done.Arguments,
					Arguments:      builder.Arguments.String(),
					Completed:      true,
				})
			}
		case "error":
			err := event.AsError()
			return ResponseResult{}, fmt.Errorf("response error %s: %s", err.Code, err.Message)
		case "response.failed":
			failed := event.AsResponseFailed()
			return ResponseResult{}, fmt.Errorf("response failed: %s", failed.Response.Error.Message)
		}
	}

	if err := stream.Err(); err != nil {
		return ResponseResult{}, err
	}

	result := ResponseResult{Text: textBuilder.String()}
	if len(order) > 0 {
		for _, itemID := range order {
			builder := builders[itemID]
			if builder == nil {
				continue
			}
			tc := ToolCall{
				ItemID:    builder.ItemID,
				CallID:    builder.CallID,
				Name:      builder.Name,
				Arguments: builder.Arguments.String(),
			}
			if tc.Arguments != "" && json.Valid([]byte(tc.Arguments)) {
				tc.ArgumentsJSON = json.RawMessage(append([]byte(nil), []byte(tc.Arguments)...))
			}
			result.ToolCalls = append(result.ToolCalls, tc)
		}
	}

	return result, nil
}

type toolCallBuilder struct {
	ItemID    string
	CallID    string
	Name      string
	Arguments strings.Builder
	Completed bool
}
