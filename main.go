package main

import (
<<<<<<< HEAD
    "context"
    "errors"
    "log"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "github.com/joho/godotenv"
    "guava/internal/openaiclient"
    "guava/internal/server"
)

type Input struct {
	client    *appopenai.Client
	prompt    string
	config    appopenai.ResponseConfig
	callbacks appopenai.StreamCallbacks
}

type Plugin struct {
}

/*
	/

/ the string to be encoded

	str := "This is an example sentence to try encoding out on!"

	result, err := encode(str)
	if err != nil {
		log.Fatalf("Encoding failed: %v", err)
	}

	// print the encoded string and token count
	fmt.Printf("Encoded tokens: %v\n", result.Tokens)
	fmt.Printf("Token count: %d\n", result.
)

*
*/

func main() {
    if err := run(); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}

func run() error {
    _ = godotenv.Load()

    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

    client, err := openaiclient.NewFromEnvironment()
    if err != nil {
        return err
    }

    srv := server.New(server.Config{
        AllowedOrigins: parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS")),
        Responses:      client.Responses(),
        Logger:         logger,
    })

    httpServer := &http.Server{
        Addr:              ":" + serverPort(),
        Handler:           srv,
        ReadTimeout:       15 * time.Second,
        ReadHeaderTimeout: 10 * time.Second,
        WriteTimeout:      60 * time.Second,
        IdleTimeout:       120 * time.Second,
    }

    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

    serverErr := make(chan error, 1)
    go func() {
        logger.Info("http server starting", slog.String("addr", httpServer.Addr))
        if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
            serverErr <- err
        }
        close(serverErr)
    }()

    select {
    case err := <-serverErr:
        if err != nil {
            return err
        }
        return nil
    case sig := <-shutdown:
        logger.Info("shutdown signal received", slog.Any("signal", sig))
    }

    ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()

    if err := httpServer.Shutdown(ctx); err != nil {
        return err
    }

    return nil
}

func parseAllowedOrigins(raw string) []string {
    if raw == "" {
        return nil
    }
    parts := strings.Split(raw, ",")
    out := make([]string, 0, len(parts))
    for _, part := range parts {
        trimmed := strings.TrimSpace(part)
        if trimmed != "" {
            out = append(out, trimmed)
        }
    }
    return out
}

func serverPort() string {
    port := os.Getenv("PORT")
    if port == "" {
        return "8080"
    }
    return port
=======
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/sashabaranov/go-openai"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: could not load .env file", err)
	}

	apiKey := os.Getenv("OPENAI_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_KEY environment variable is not set")
	}

	client := openai.NewClient(apiKey)
	wrapper := NewOpenAIWrapper(client)
	server := NewAPIServer(wrapper)

	router := mux.NewRouter()
	router.HandleFunc("/v1/responses", server.handleCreateResponse).Methods(http.MethodPost)
	router.HandleFunc("/v1/threads", server.handleCreateThread).Methods(http.MethodPost)
	router.HandleFunc("/v1/assistants", server.handleCreateAssistant).Methods(http.MethodPost)
	router.HandleFunc("/v1/vector-stores", server.handleCreateVectorStore).Methods(http.MethodPost)
	router.HandleFunc("/openapi.json", server.handleOpenAPIDocument).Methods(http.MethodGet)
	router.HandleFunc("/swagger.json", server.handleOpenAPIDocument).Methods(http.MethodGet)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}).Handler(router)

	log.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
>>>>>>> origin/refactor-api-for-versioned-dtos-and-handlers
}
