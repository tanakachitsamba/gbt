package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	gogpt "github.com/sashabaranov/go-gpt3"
)

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

func main() {
	var err error = godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file: ", err)
	}

	g := gogpt.NewClient(os.Getenv("OPENAI_KEY"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "POST" {
			// Get the data from the form
			userInput := r.FormValue("input")
			sessionID := r.FormValue("sessionID")

			// Get the data from the database
			//var taskTodo = []string{``}
			// ...

			// Process the data
			var (
				res        string
				promptList = []string{}
			)
			res, promptList = bot(userInput, promptList, g)
			_, _ = promptList, res

			// Store the data in the database
			// ...

			// Render the template
			t, _ := template.ParseFiles("template.html")
			t.Execute(w, map[string]interface{}{
				"Input":    userInput,
				"BotReply": res,
				"Session":  sessionID,
			})
		} else {
			// Render the template
			t, _ := template.ParseFiles("template.html")
			t.Execute(w, nil)
		}
	})

	http.ListenAndServe(":8080", nil)
}
