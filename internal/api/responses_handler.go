package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/openai/openai-go/responses"

	"guava/internal/openaiclient"
)

// ResponsesHandler exposes HTTP endpoints for working with the OpenAI Responses
// API.
type ResponsesHandler struct {
	responses openaiclient.ResponsesClient
	logger    *slog.Logger
}

// NewResponsesHandler constructs a new handler. A nil logger results in the
// default slog logger being used.
func NewResponsesHandler(client openaiclient.ResponsesClient, logger *slog.Logger) *ResponsesHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &ResponsesHandler{responses: client, logger: logger}
}

// HandleCreateResponse accepts POST requests to create a new model response. If
// the request payload sets stream=true the response is streamed using
// server-sent events (SSE).
func (h *ResponsesHandler) HandleCreateResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST, OPTIONS")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req CreateResponseRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON payload: %v", err))
		return
	}

	if err := req.Validate(); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	params, err := BuildResponseParams(&req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Stream {
		h.streamResponse(w, r, params)
		return
	}

	result, err := h.responses.CreateResponse(r.Context(), params)
	if err != nil {
		h.logger.Error("create response", slog.String("error", err.Error()))
		writeJSONError(w, http.StatusBadGateway, "failed to create response")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *ResponsesHandler) streamResponse(w http.ResponseWriter, r *http.Request, params responses.ResponseNewParams) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming not supported by server")
		return
	}

	stream, err := h.responses.StreamResponse(r.Context(), params)
	if err != nil {
		h.logger.Error("create streaming response", slog.String("error", err.Error()))
		writeJSONError(w, http.StatusBadGateway, "failed to create response stream")
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var aggregated strings.Builder

	for stream.Next() {
		event := stream.Current()
		payload, include := transformStreamEvent(event, &aggregated)
		if !include {
			continue
		}
		if err := writeSSE(w, flusher, payload); err != nil {
			h.logger.Error("write SSE", slog.String("error", err.Error()))
			return
		}
	}

	if err := stream.Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}
		if ctxErr := r.Context().Err(); errors.Is(ctxErr, context.Canceled) || errors.Is(ctxErr, context.DeadlineExceeded) {
			return
		}

		h.logger.Error("stream error", slog.String("error", err.Error()))
		_ = writeSSE(w, flusher, StreamMessage{
			Type:  "stream.error",
			Error: &StreamError{Message: err.Error()},
		})
	}
}

// StreamMessage represents a single SSE payload returned to the client.
type StreamMessage struct {
	Type           string              `json:"type"`
	ItemID         string              `json:"item_id,omitempty"`
	OutputIndex    *int64              `json:"output_index,omitempty"`
	ContentIndex   *int64              `json:"content_index,omitempty"`
	Delta          string              `json:"delta,omitempty"`
	Text           string              `json:"text,omitempty"`
	AggregatedText string              `json:"aggregated_text,omitempty"`
	Response       *responses.Response `json:"response,omitempty"`
	Error          *StreamError        `json:"error,omitempty"`
}

// StreamError contains details when streaming fails mid-flight.
type StreamError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}

func transformStreamEvent(event responses.ResponseStreamEventUnion, aggregate *strings.Builder) (StreamMessage, bool) {
	switch event.Type {
	case "response.output_text.delta":
		delta := event.AsResponseOutputTextDelta()
		aggregate.WriteString(delta.Delta)
		return StreamMessage{
			Type:         event.Type,
			ItemID:       delta.ItemID,
			OutputIndex:  &delta.OutputIndex,
			ContentIndex: &delta.ContentIndex,
			Delta:        delta.Delta,
		}, true
	case "response.output_text.done":
		done := event.AsResponseOutputTextDone()
		return StreamMessage{
			Type:         event.Type,
			ItemID:       done.ItemID,
			OutputIndex:  &done.OutputIndex,
			ContentIndex: &done.ContentIndex,
			Text:         done.Text,
		}, true
	case "response.completed":
		completed := event.AsResponseCompleted()
		return StreamMessage{
			Type:           event.Type,
			Response:       &completed.Response,
			AggregatedText: aggregate.String(),
		}, true
	case "response.failed":
		failed := event.AsResponseFailed()
		return StreamMessage{
			Type:     event.Type,
			Response: &failed.Response,
		}, true
	case "error":
		errEvent := event.AsError()
		return StreamMessage{
			Type: event.Type,
			Error: &StreamError{
				Code:    string(errEvent.Code),
				Message: errEvent.Message,
				Param:   errEvent.Param,
			},
		}, true
	default:
		return StreamMessage{}, false
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, payload StreamMessage) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}
	// json.Encoder includes a trailing newline; we strip it to maintain SSE
	// formatting semantics.
	data := strings.TrimSpace(buf.String())
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: ErrorBody{Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// The response writer is already committed, so the best we can do is log.
		slog.Default().Error("write JSON", slog.String("error", err.Error()))
	}
}
