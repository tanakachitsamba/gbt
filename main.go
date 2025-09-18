package main

import (
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
}
