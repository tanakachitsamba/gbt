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
	var err error = godotenv.Load(".env")
	if err != nil {
		log.Println("error loading .env file:", err)
	}

	//i := interpreter("what is the best way to make a pizza?")
	//log.Println(i)
	//x := getResponse(i)
	//log.Println(x)

	instruct := `create a report on how this agent could be useful.`
	res := runPolicies(instruct)
	_ = res

	log.Println(res)

	// sending sms
	//sendSMS(os.Getenv("TWILIO_PHONE_FROM"), os.Getenv("TWILIO_PHONE_TO"), quote)
}

func getResponse(prompt string) string {
	// Get the OpenAI API key from the .env file
	if err := godotenv.Load(); err != nil {
		log.Println("error loading .env file:", err)
	}

	g := gogpt.NewClient(os.Getenv("OPENAI_KEY"))

	var (
		botResponse string
		err         error
	)
	_ = err

	botResponse, err = getChatStreamResponse(prompt, g, 250)
	return botResponse

}

func getChatStreamResponse(prompt string, g *gogpt.Client, maxTokens int) (string, error) {
	request := gogpt.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []gogpt.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:       maxTokens,
		Temperature:     0,
		TopP:            1,
		PresencePenalty: 0.6,
		Stop:            []string{"tanaka:", "enquirer:", "reflector:", "prioritiser:", "planner:", "lister:", "decider:", "policy-decider:", "criticiser:", "recaller:", "tokensniffer:", "host:"},
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
			break
		}
	}

	return buffer.String(), nil
}

func getStreamResponse(prompt string, g *gogpt.Client) (string, error) {
	request := gogpt.CompletionRequest{
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
