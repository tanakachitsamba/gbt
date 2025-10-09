package server

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	"github.com/rs/cors"

	"guava/internal/api"
	"guava/internal/openaiclient"
	"guava/internal/prompt"
)

// Config configures the HTTP server handler.
type Config struct {
	AllowedOrigins []string
	Responses      openaiclient.ResponsesClient
	Logger         *slog.Logger
	DuckDB         *sql.DB
	PromptService  prompt.Evaluator
	PromptModel    string
}

// New constructs an HTTP handler configured with the provided options. The
// handler registers endpoints for health checks and the Responses API.
func New(cfg Config) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	mux := http.NewServeMux()
	responsesHandler := api.NewResponsesHandler(cfg.Responses, cfg.Logger)
	mux.HandleFunc("/v1/responses", responsesHandler.HandleCreateResponse)

	evaluator := cfg.PromptService
	if evaluator == nil && cfg.DuckDB != nil && cfg.Responses != nil {
		repo, err := prompt.NewDuckDBRepository(cfg.DuckDB)
		if err != nil {
			cfg.Logger.Error("initialise prompt repository", slog.String("error", err.Error()))
		} else {
			service, svcErr := prompt.NewService(prompt.ServiceConfig{
				Responses:      cfg.Responses,
				Repository:     repo,
				EvaluatorModel: cfg.PromptModel,
				Logger:         cfg.Logger,
			})
			if svcErr != nil {
				cfg.Logger.Error("initialise prompt service", slog.String("error", svcErr.Error()))
			} else {
				evaluator = service
			}
		}
	}
	if evaluator != nil {
		evaluationHandler := api.NewPromptEvaluationHandler(evaluator, cfg.Logger)
		mux.HandleFunc("/v1/prompt-evaluations", evaluationHandler.HandleEvaluatePrompt)
	}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	allowed := cfg.AllowedOrigins
	if len(allowed) == 0 {
		allowed = []string{"http://localhost:3000"}
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   allowed,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	})

	// Ensure requests to unknown routes return 404 with JSON payload.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/") && r.URL.Path != "/healthz" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"route not found"}}`))
			return
		}
		mux.ServeHTTP(w, r)
	})

	return corsHandler.Handler(handler)
}
