package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/sashabaranov/go-openai"
	//"github.com/joho/godotenv"
	//gogpt "github.com/sashabaranov/go-gpt3"
)

type Input struct {
	client                       *openai.Client
	prompt, model, systemMessage string
	temperature                  float32
	maxTokens                    int
	res                          string // maybe this could be a generic so that it can be both a slice, string or a null
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

func getClient() *openai.Client {
	// Get the OpenAI API key from the .env file
	if err := godotenv.Load(); err != nil {
		log.Println("error loading .env file:", err)
	}

	var key string = os.Getenv("OPENAI_KEY")

	return openai.NewClient(key)
}

func (inp Input) getChatStreamResponse() (string, error) {
	var array = []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: inp.systemMessage,
		},
		{
			Role:    "user",
			Content: inp.prompt,
		},
	}

	request := openai.ChatCompletionRequest{
		Model:           inp.model,
		Messages:        array,
		MaxTokens:       inp.maxTokens,
		Temperature:     inp.temperature,
		TopP:            1,
		PresencePenalty: 0.6,
		Stop:            []string{"user:", "assistant:"},
	}

	stream, err := inp.client.CreateChatCompletionStream(context.Background(), request)
	if err != nil {
		log.Println(err, "createchatcompletionstream")
		return "", err
	}
	defer stream.Close()

	var buffer strings.Builder
	for {
		response, err := stream.Recv()
		if err != nil {
			log.Println(err, "stream.Recv()")
			return "", err
		}

		if len(response.Choices) > 0 {
			choice := response.Choices[0]
			buffer.WriteString(choice.Delta.Content)
		}

		if response.Choices[0].FinishReason != "" {
			break
		}
	}

	return buffer.String(), nil
}

func getStreamResponse(prompt string, g *openai.Client) (string, error) {
	request := openai.CompletionRequest{
		Model:     "text-ada-001",
		MaxTokens: 500,
		Prompt:    prompt,
		Stream:    true,
		//Stop:            []string{"human:", "ai:"},
		//Temperature:     0,
		//TopP:            1,
		//PresencePenalty: 0.6,
	}

	stream, err := g.CreateCompletionStream(context.Background(), request)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	var buffer strings.Builder
	for {
		response, err := stream.Recv()
		if err != nil {
			return "", err
		}

		buffer.WriteString(response.Choices[0].Text)

		if response.Choices[0].FinishReason != "" {
			break
		}
	}

	return buffer.String(), nil
}
