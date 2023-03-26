package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	gogpt "github.com/sashabaranov/go-gpt3"
)

func main() {
	http.HandleFunc("/input", handleInput)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type InputRequest struct {
	Message string `json:"message"`
}

type OutputResponse struct {
	Response string `json:"response"`
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
		MaxTokens:       150,
		Temperature:     1,
		TopP:            1,
		PresencePenalty: 0.6,
		Stop:            []string{"Human:", "AI:"},
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

func updateList(message string, pl []string) []string {
	return append(pl, message)
}

func createPrompt(message string, pl []string) string {
	pMessage := fmt.Sprintf("\nHuman: %s", message)
	prompt := append(pl, pMessage)
	return strings.Join(prompt, "")
}

// what is this doing?
func findBotMessagePosition(response string) int {
	// Find the position of the last occurrence of "\nHuman:"
	humanPos := strings.LastIndex(response, "\nHuman:")
	if humanPos == -1 {
		// No "\nHuman:" found, so assume the entire response is the bot message
		return 0
	}

	// Find the position of the next "\nAI:" after "\nHuman:"
	aiPos := strings.Index(response[humanPos:], "\nAI:")
	if aiPos == -1 {
		// No "\nAI:" found after "\nHuman:", so assume the bot message starts at the end of the "\nHuman:" line
		return humanPos + len("\nHuman:")
	}
	// Return the position of the start of the bot message
	return humanPos + aiPos + len("\nAI:")
}

func getBotResponse(prompt string, g *gogpt.Client) (string, error) {
	botResponse, err := getStreamResponse(prompt, g)
	if err != nil {
		return "", err
	}

	pos := findBotMessagePosition(botResponse)
	if pos != -1 {
		botResponse = botResponse[pos:]
	} else {
		botResponse = "Something went wrong..."
	}
	return botResponse, nil
}

var bot = func(input string, promptList []string, g *gogpt.Client) (string, []string) {
	input = createPrompt(input, promptList)
	response, err := getBotResponse(input, g)
	if err != nil {
		fmt.Printf("Error getting bot response: %v\n", err)
	}
	log.Println(response)
	promptList = updateList(response, promptList)
	return response, promptList
}

func handleInput(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var inputReq InputRequest
	err := json.NewDecoder(r.Body).Decode(&inputReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	g := gogpt.NewClient(os.Getenv("OPENAI_KEY"))
	var promptList = []string{}
	response, promptList := bot(inputReq.Message, promptList, g)
	_ = promptList

	outputRes := OutputResponse{Response: response}
	err = json.NewEncoder(w).Encode(&outputRes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("Received input:", inputReq.Message)
	log.Println("Processed response:", response)
}
