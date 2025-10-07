package api

import (
	"errors"
	"strings"
	"time"
)

// PromptEvaluationRequest captures the payload for requesting a prompt critique
// and improvement.
type PromptEvaluationRequest struct {
	Prompt          string                 `json:"prompt"`
	TargetModel     string                 `json:"target_model"`
	EvaluationModel string                 `json:"evaluation_model,omitempty"`
	ResearchTopics  []string               `json:"research_topics,omitempty"`
	MaxContext      int                    `json:"max_context,omitempty"`
	Backend         *PromptBackendOverride `json:"backend,omitempty"`
	Metadata        map[string]string      `json:"metadata,omitempty"`
}

// PromptBackendOverride mirrors the backend override options exposed by the
// prompt service allowing callers to direct evaluations to alternative model
// endpoints.
type PromptBackendOverride struct {
	BaseURL string            `json:"base_url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Validate performs minimal validation on the request payload.
func (r *PromptEvaluationRequest) Validate() error {
	if r == nil {
		return errors.New("request body is required")
	}
	if strings.TrimSpace(r.Prompt) == "" {
		return errors.New("prompt is required")
	}
	if strings.TrimSpace(r.TargetModel) == "" {
		return errors.New("target_model is required")
	}
	if r.MaxContext < 0 {
		return errors.New("max_context must be zero or positive")
	}
	if r.Backend != nil {
		for key := range r.Backend.Headers {
			if strings.TrimSpace(key) == "" {
				return errors.New("backend.headers keys must be non-empty")
			}
		}
	}
	return nil
}

// PromptEvaluationResponse returns the improved prompt and evaluation metadata.
type PromptEvaluationResponse struct {
	EvaluationID    string                  `json:"evaluation_id"`
	TargetModel     string                  `json:"target_model"`
	EvaluationModel string                  `json:"evaluation_model"`
	OriginalPrompt  string                  `json:"original_prompt"`
	ImprovedPrompt  string                  `json:"improved_prompt"`
	Critique        string                  `json:"critique,omitempty"`
	Scores          map[string]float64      `json:"scores"`
	Suggestions     []string                `json:"suggestions,omitempty"`
	ResearchContext []PromptResearchContext `json:"research_context"`
	RawModelOutput  string                  `json:"raw_model_output"`
	Metadata        map[string]string       `json:"metadata,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
}

// PromptResearchContext mirrors the repository research context for API
// responses.
type PromptResearchContext struct {
	ID          string    `json:"id"`
	TargetModel string    `json:"target_model,omitempty"`
	Topic       string    `json:"topic,omitempty"`
	Content     string    `json:"content"`
	Source      string    `json:"source,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
