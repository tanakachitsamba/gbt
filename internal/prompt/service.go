package prompt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"

	"guava/internal/openaiclient"
)

var (
	// ErrMissingResponses indicates the service cannot operate without a
	// Responses client.
	ErrMissingResponses = errors.New("prompt service requires responses client")
	// ErrMissingRepository indicates the service cannot operate without a
	// repository implementation.
	ErrMissingRepository = errors.New("prompt service requires repository")
	// ErrInvalidInput is returned when the caller supplies an invalid request.
	ErrInvalidInput = errors.New("invalid prompt evaluation input")
)

// Service orchestrates prompt evaluations by combining stored research context,
// the OpenAI Responses API, and heuristics for scoring improvements.
type Service struct {
	responses      openaiclient.ResponsesClient
	repository     Repository
	evaluatorModel string
	logger         *slog.Logger
}

// BackendOverride allows callers to route evaluations to alternative model
// backends such as OSS deployments by overriding the base URL or supplying
// additional HTTP headers.
type BackendOverride struct {
	BaseURL string
	Headers map[string]string
}

// ServiceConfig configures a new Service instance.
type ServiceConfig struct {
	Responses      openaiclient.ResponsesClient
	Repository     Repository
	EvaluatorModel string
	Logger         *slog.Logger
}

// EvaluationInput captures the caller supplied data used when critiquing a
// prompt.
type EvaluationInput struct {
	Prompt          string
	TargetModel     string
	EvaluationModel string
	ResearchTopics  []string
	MaxContext      int
	Backend         *BackendOverride
	Metadata        map[string]string
}

// EvaluationResult contains the improved prompt along with metadata describing
// the evaluation run and supporting context.
type EvaluationResult struct {
	EvaluationID    string
	TargetModel     string
	EvaluationModel string
	OriginalPrompt  string
	ImprovedPrompt  string
	Critique        string
	Scores          map[string]float64
	Suggestions     []string
	ResearchContext []ResearchContext
	RawModelOutput  string
	Metadata        map[string]string
	CreatedAt       time.Time
}

// Evaluator exposes the EvaluatePrompt method implemented by Service. It is
// defined to simplify testing/mocking at the HTTP layer.
type Evaluator interface {
	EvaluatePrompt(ctx context.Context, input EvaluationInput) (*EvaluationResult, error)
}

// NewService constructs a new evaluator instance.
func NewService(cfg ServiceConfig) (*Service, error) {
	if cfg.Responses == nil {
		return nil, ErrMissingResponses
	}
	if cfg.Repository == nil {
		return nil, ErrMissingRepository
	}
	srv := &Service{
		responses:      cfg.Responses,
		repository:     cfg.Repository,
		evaluatorModel: cfg.EvaluatorModel,
		logger:         cfg.Logger,
	}
	if srv.logger == nil {
		srv.logger = slog.Default()
	}
	return srv, nil
}

// EvaluatePrompt critiques the provided prompt, improves it using the configured
// model, and returns the enriched prompt along with scoring metadata.
func (s *Service) EvaluatePrompt(ctx context.Context, input EvaluationInput) (*EvaluationResult, error) {
	if strings.TrimSpace(input.Prompt) == "" {
		return nil, fmt.Errorf("%w: prompt is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.TargetModel) == "" {
		return nil, fmt.Errorf("%w: target_model is required", ErrInvalidInput)
	}

	evaluationModel := input.EvaluationModel
	if evaluationModel == "" {
		evaluationModel = s.evaluatorModel
	}
	if evaluationModel == "" {
		return nil, fmt.Errorf("%w: evaluation model is not configured", ErrInvalidInput)
	}

	contexts, err := s.repository.ListResearchContext(ctx, input.TargetModel, input.ResearchTopics, input.MaxContext)
	if err != nil {
		return nil, fmt.Errorf("fetch research context: %w", err)
	}

	instructions := buildInstructions(input.TargetModel)
	payload := buildModelInput(input.Prompt, input.TargetModel, contexts, input.Metadata)

	params := responses.ResponseNewParams{
		Model:        responses.ResponsesModel(evaluationModel),
		Input:        responses.ResponseNewParamsInputUnion{OfString: openai.String(payload)},
		Instructions: openai.String(instructions),
	}

	opts := buildRequestOptions(input.Backend)

	result, err := s.responses.CreateResponseWithOptions(ctx, params, opts...)
	if err != nil {
		return nil, fmt.Errorf("create evaluation response: %w", err)
	}

	feedback, err := decodeModelFeedback(result.OutputText)
	if err != nil {
		return nil, fmt.Errorf("parse evaluation feedback: %w", err)
	}

	if feedback.ImprovedPrompt == "" {
		return nil, fmt.Errorf("%w: model did not return an improved_prompt", ErrInvalidInput)
	}

	scores := mergeScores(input.Prompt, feedback, contexts)

	createdAt := time.Now().UTC()
	evaluationID := uuid.NewString()

	record := PromptEvaluationRecord{
		EvaluationID:    evaluationID,
		TargetModel:     input.TargetModel,
		EvaluationModel: evaluationModel,
		OriginalPrompt:  input.Prompt,
		ImprovedPrompt:  feedback.ImprovedPrompt,
		Critique:        feedback.Critique,
		Scores:          scores,
		Suggestions:     feedback.Suggestions,
		RawModelOutput:  result.OutputText,
		Metadata:        input.Metadata,
		CreatedAt:       createdAt,
	}
	record.References = make([]ResearchReference, 0, len(contexts))
	for _, ctxItem := range contexts {
		record.References = append(record.References, ResearchReference{
			ContextID: ctxItem.ID,
			Topic:     ctxItem.Topic,
			Content:   ctxItem.Content,
			Source:    ctxItem.Source,
		})
	}

	if err := s.repository.SavePromptEvaluation(ctx, record); err != nil {
		return nil, fmt.Errorf("persist evaluation: %w", err)
	}

	return &EvaluationResult{
		EvaluationID:    evaluationID,
		TargetModel:     input.TargetModel,
		EvaluationModel: evaluationModel,
		OriginalPrompt:  input.Prompt,
		ImprovedPrompt:  feedback.ImprovedPrompt,
		Critique:        feedback.Critique,
		Scores:          scores,
		Suggestions:     feedback.Suggestions,
		ResearchContext: contexts,
		RawModelOutput:  result.OutputText,
		Metadata:        input.Metadata,
		CreatedAt:       createdAt,
	}, nil
}

type modelFeedback struct {
	ImprovedPrompt string             `json:"improved_prompt"`
	Critique       string             `json:"critique"`
	Suggestions    []string           `json:"suggestions"`
	Scores         map[string]float64 `json:"scores"`
}

func buildInstructions(targetModel string) string {
	var sb strings.Builder
	sb.WriteString("You are an expert prompt engineer. Improve the supplied prompt for the target model \"")
	sb.WriteString(targetModel)
	sb.WriteString("\" while preserving its intent. Respond strictly in JSON with the keys improved_prompt, critique, suggestions (array), and scores (object).")
	return sb.String()
}

func buildModelInput(prompt, targetModel string, contexts []ResearchContext, metadata map[string]string) string {
	var sb strings.Builder
	sb.WriteString("Target Model: ")
	sb.WriteString(targetModel)
	sb.WriteString("\n")
	sb.WriteString("Original Prompt:\n")
	sb.WriteString(prompt)
	sb.WriteString("\n\n")

	if len(contexts) > 0 {
		sb.WriteString("Research Context:\n")
		for i, ctxItem := range contexts {
			sb.WriteString(fmt.Sprintf("[%d] Topic: %s\n", i+1, ctxItem.Topic))
			sb.WriteString(ctxItem.Content)
			if ctxItem.Source != "" {
				sb.WriteString("\nSource: ")
				sb.WriteString(ctxItem.Source)
			}
			sb.WriteString("\n---\n")
		}
	}

	if len(metadata) > 0 {
		keys := make([]string, 0, len(metadata))
		for k := range metadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sb.WriteString("Metadata:\n")
		for _, key := range keys {
			sb.WriteString(key)
			sb.WriteString(": ")
			sb.WriteString(metadata[key])
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nPlease return the JSON response now.")
	return sb.String()
}

func buildRequestOptions(backend *BackendOverride) []option.RequestOption {
	if backend == nil {
		return nil
	}
	opts := make([]option.RequestOption, 0, 1+len(backend.Headers))
	if backend.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(backend.BaseURL))
	}
	for key, value := range backend.Headers {
		if key == "" {
			continue
		}
		opts = append(opts, option.WithHeader(key, value))
	}
	return opts
}

func decodeModelFeedback(raw string) (*modelFeedback, error) {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var feedback modelFeedback
	if err := json.Unmarshal([]byte(cleaned), &feedback); err != nil {
		return nil, err
	}
	if feedback.Scores == nil {
		feedback.Scores = make(map[string]float64)
	}
	return &feedback, nil
}

var wordCleaner = regexp.MustCompile(`[^a-z0-9]+`)

func mergeScores(originalPrompt string, feedback *modelFeedback, contexts []ResearchContext) map[string]float64 {
	scores := make(map[string]float64, len(feedback.Scores)+4)
	for k, v := range feedback.Scores {
		scores[k] = clamp(v)
	}

	structureScore := computeStructureScore(feedback.ImprovedPrompt)
	if _, ok := scores["structure"]; !ok {
		scores["structure"] = structureScore
	}

	specificityScore := computeSpecificityScore(originalPrompt, feedback.ImprovedPrompt)
	if _, ok := scores["specificity"]; !ok {
		scores["specificity"] = specificityScore
	}

	supportScore := computeSupportScore(feedback.Suggestions)
	if _, ok := scores["support"]; !ok {
		scores["support"] = supportScore
	}

	contextScore := computeContextScore(contexts)
	if _, ok := scores["context_alignment"]; !ok {
		scores["context_alignment"] = contextScore
	}

	if _, ok := scores["overall"]; !ok {
		sum := 0.0
		for _, v := range scores {
			sum += clamp(v)
		}
		if len(scores) > 0 {
			scores["overall"] = clamp(sum / float64(len(scores)))
		} else {
			scores["overall"] = 0
		}
	}

	return scores
}

func computeStructureScore(prompt string) float64 {
	lines := strings.Split(prompt, "\n")
	if len(lines) == 0 {
		return 0
	}
	bulletCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			bulletCount++
			continue
		}
		if len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9' && (trimmed[1] == '.' || trimmed[1] == ')') {
			bulletCount++
		}
	}
	ratio := float64(bulletCount+1) / float64(len(lines)+1)
	return clamp(ratio * 1.2)
}

func computeSpecificityScore(original, improved string) float64 {
	originalWords := uniqueWords(original)
	improvedWords := uniqueWords(improved)
	if len(improvedWords) == 0 {
		return 0
	}
	overlap := 0
	for word := range improvedWords {
		if _, ok := originalWords[word]; !ok {
			overlap++
		}
	}
	ratio := float64(overlap) / float64(len(improvedWords))
	return clamp(0.5 + ratio/2)
}

func computeSupportScore(suggestions []string) float64 {
	if len(suggestions) == 0 {
		return 0.5
	}
	ratio := float64(len(suggestions)) / 5.0
	return clamp(ratio)
}

func computeContextScore(contexts []ResearchContext) float64 {
	if len(contexts) == 0 {
		return 0.3
	}
	length := 0
	for _, ctx := range contexts {
		length += len(ctx.Content)
	}
	ratio := float64(length) / 2000.0
	return clamp(0.6 + math.Min(0.4, ratio))
}

func uniqueWords(input string) map[string]struct{} {
	words := strings.Fields(strings.ToLower(input))
	out := make(map[string]struct{}, len(words))
	for _, word := range words {
		cleaned := wordCleaner.ReplaceAllString(word, "")
		if cleaned == "" {
			continue
		}
		out[cleaned] = struct{}{}
	}
	return out
}

func clamp(value float64) float64 {
	switch {
	case math.IsNaN(value):
		return 0
	case value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}
