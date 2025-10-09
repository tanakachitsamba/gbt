package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"guava/internal/prompt"
)

// PromptEvaluationHandler exposes the HTTP surface for evaluating prompts.
type PromptEvaluationHandler struct {
	evaluator prompt.Evaluator
	logger    *slog.Logger
}

// NewPromptEvaluationHandler constructs a handler with the provided evaluator.
func NewPromptEvaluationHandler(evaluator prompt.Evaluator, logger *slog.Logger) *PromptEvaluationHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &PromptEvaluationHandler{evaluator: evaluator, logger: logger}
}

// HandleEvaluatePrompt accepts POST requests for /v1/prompt-evaluations.
func (h *PromptEvaluationHandler) HandleEvaluatePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST, OPTIONS")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.evaluator == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "prompt evaluation service not configured")
		return
	}

	var req PromptEvaluationRequest
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

	input := prompt.EvaluationInput{
		Prompt:          req.Prompt,
		TargetModel:     req.TargetModel,
		EvaluationModel: req.EvaluationModel,
		ResearchTopics:  req.ResearchTopics,
		MaxContext:      req.MaxContext,
		Metadata:        req.Metadata,
	}
	if req.Backend != nil {
		input.Backend = &prompt.BackendOverride{
			BaseURL: req.Backend.BaseURL,
			Headers: req.Backend.Headers,
		}
	}

	result, err := h.evaluator.EvaluatePrompt(r.Context(), input)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, prompt.ErrInvalidInput) {
			status = http.StatusBadRequest
		}
		h.logger.Error("evaluate prompt", slog.String("error", err.Error()))
		writeJSONError(w, status, "failed to evaluate prompt")
		return
	}

	resp := PromptEvaluationResponse{
		EvaluationID:    result.EvaluationID,
		TargetModel:     result.TargetModel,
		EvaluationModel: result.EvaluationModel,
		OriginalPrompt:  result.OriginalPrompt,
		ImprovedPrompt:  result.ImprovedPrompt,
		Critique:        result.Critique,
		Scores:          result.Scores,
		Suggestions:     result.Suggestions,
		RawModelOutput:  result.RawModelOutput,
		Metadata:        result.Metadata,
		CreatedAt:       result.CreatedAt,
	}
	resp.ResearchContext = make([]PromptResearchContext, 0, len(result.ResearchContext))
	for _, ctxItem := range result.ResearchContext {
		resp.ResearchContext = append(resp.ResearchContext, PromptResearchContext{
			ID:          ctxItem.ID,
			TargetModel: ctxItem.TargetModel,
			Topic:       ctxItem.Topic,
			Content:     ctxItem.Content,
			Source:      ctxItem.Source,
			CreatedAt:   ctxItem.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}
