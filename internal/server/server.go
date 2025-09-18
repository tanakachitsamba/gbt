package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/rs/cors"

	"guava/internal/api"
	"guava/internal/openaiclient"
)

// Config configures the HTTP server handler.
type Config struct {
	AllowedOrigins []string
	Responses      openaiclient.ResponsesClient
	Logger         *slog.Logger
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
