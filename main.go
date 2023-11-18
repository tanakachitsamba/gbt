package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/sashabaranov/go-openai"
	//"github.com/joho/godotenv"
	//gogpt "github.com/sashabaranov/go-gpt3"
)

type Input struct {
	client                       *openai.Client
	prompt, model, systemMessage string
	temperature                  float32
	maxTokens                    int
}

func main() {
	inp := Input{client: getClient(), prompt: "Create questions for a job interview for a financial accounts manager in the uk" + "\n", model: "gpt-3.5-turbo-0613", temperature: 0.8, maxTokens: 1000, systemMessage: `Don't make assumptions about what values to plug into functions. Ask for clarification if a user request is ambiguous`}

	_, err, thing := inp.getChatStreamResponse()
	_ = thing

	if err != nil {
		log.Println(err, "error from the model")
	}
}

func getClient() *openai.Client {
	// Get the OpenAI API key from the .env file
	if err := godotenv.Load(); err != nil {
		log.Println("error loading .env file:", err)
	}

	var key string = os.Getenv("OPENAI_KEY")

	return openai.NewClient(key)
}

func (inp Input) getChatStreamResponse() (string, error, string) {
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

	var request openai.ChatCompletionRequest = openai.ChatCompletionRequest{
		Model:           inp.model,
		Messages:        array,
		MaxTokens:       inp.maxTokens,
		Temperature:     inp.temperature,
		TopP:            1,
		PresencePenalty: 0.6,
		Stop:            []string{"user:", "assistant:"},
		Functions: []openai.FunctionDefinition{
			{
				Name:        "interview_questions",
				Description: "gets questions for a job interview for a financial accounts manager in the UK in an array",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"questions": map[string]interface{}{
							"type":        "array",
							"items":       map[string]interface{}{"type": "string"},
							"description": "List of interview questions",
						},
					},
					"required": []string{"questions"},
				},
			},
		},
		FunctionCall: map[string]interface{}{"name": "interview_questions"},
	}

	stream, err := inp.client.CreateChatCompletionStream(context.Background(), request)
	if err != nil {
		log.Println(err, "createchatcompletionstream")
		return "", err, ""
	}
	defer stream.Close()

	var buffer strings.Builder
	var args []string
	for {
		response, err := stream.Recv()
		if err != nil {
			log.Println(err, "stream.Recv()")
			return "", err, ""
		}

		if len(response.Choices) > 0 {
			choice := response.Choices[0]
			buffer.WriteString(choice.Delta.Content)

			if choice.Delta.FunctionCall != nil {
				args = append(args, choice.Delta.FunctionCall.Arguments)
			}

		}

		if response.Choices[0].FinishReason != "" {

			break
		}

	}

	res := strings.Join(args, "")
	log.Println(res)

	return buffer.String(), nil, res
}
