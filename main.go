package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/openai/openai-go/option"
	appopenai "guava/pkg/openai"
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
	// Enable CORS with allowed origins
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"POST"},
		AllowedHeaders: []string{"Content-Type"},
	})

	handler := c.Handler(http.HandlerFunc(handleRequest))

	log.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func getClient() *appopenai.Client {
	// Get the OpenAI API key from the .env file
	if err := godotenv.Load(); err != nil {
		log.Println("error loading .env file:", err)
	}

	key := os.Getenv("OPENAI_KEY")
	if key == "" {
		return appopenai.NewClient()
	}

	return appopenai.NewClient(option.WithAPIKey(key))
}

func (inp Input) getChatStreamResponse() (string, error) {
	if inp.client == nil {
		return "", appopenai.ErrMissingClient
	}

	result, err := inp.client.StreamResponse(context.Background(), appopenai.ResponseRequest{
		Input:     inp.prompt,
		Config:    inp.config,
		Callbacks: inp.callbacks,
	})
	if err != nil {
		log.Println(err, "responses stream")
		return "", err
	}

	return result.Text, nil
}

func getStreamResponse(prompt string, client *appopenai.Client) (string, error) {
	if client == nil {
		return "", appopenai.ErrMissingClient
	}

	result, err := client.StreamResponse(context.Background(), appopenai.ResponseRequest{
		Input: prompt,
		Config: appopenai.ResponseConfig{
			Model: appopenai.ModelGPT4oMini,
		},
	})
	if err != nil {
		return "", err
	}

	return result.Text, nil
}
