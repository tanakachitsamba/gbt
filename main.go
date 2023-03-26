package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	gogpt "github.com/sashabaranov/go-gpt3"
)

func main() {
	x := getAnswer(interpreter("can you tell me who the president of the italy is"))
	log.Println(x)
}

func getAnswer(prompt string) string {
	// Get the OpenAI API key from the .env file
	if err := godotenv.Load(); err != nil {
		log.Println("error loading .env file:", err)
	}

	g := gogpt.NewClient(os.Getenv("OPENAI_KEY"))
	botResponse, err := getStreamResponse(prompt, g)
	if err != nil {
		log.Println("error getting bot response:", err)
	}

	return botResponse

}

func getStreamResponse(prompt string, g *gogpt.Client) (string, error) {
	request := gogpt.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []gogpt.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:       250,
		Temperature:     0,
		TopP:            1,
		PresencePenalty: 0.6,
		Stop:            []string{"human:", "ai:"},
	}

	stream, err := g.CreateChatCompletionStream(context.Background(), request)
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

		if len(response.Choices) > 0 {
			choice := response.Choices[0]
			buffer.WriteString(choice.Delta.Content)
		}

		if response.Choices[0].FinishReason != "" {
			//log.Println("finish reason:", response.Choices[0].FinishReason)
			break
		}
	}

	return buffer.String(), nil
}
