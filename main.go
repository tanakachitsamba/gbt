package main

import (
	"context"
	"html"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/sashabaranov/go-openai"
	//"github.com/joho/godotenv"
	//gogpt "github.com/sashabaranov/go-gpt3"
)

type Input struct {
	client        *openai.Client
	prompt, model string
	temperature   float32
	maxTokens     int
}

type Plugin struct {
}

func stringToHTML(input string) string {
	return html.EscapeString(input)
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
	r := mux.NewRouter()
	r.HandleFunc("/", handleRequest).Methods("POST")

	// Enable CORS
	corsHandler := cors.Default().Handler(r)

	log.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))

	// this creates files
	/*
		// check if output exist and if it does it gets moved to history under a unique name
		err = ioutil.WriteFile("output.txt", []byte(str), 0644)
		if err != nil {
			panic(err)
		}

	*/

	/*
			fileName := "app.html"

		err = WriteHTMLFile(fileName, str)
		if err != nil {
			fmt.Println("Error writing HTML file:", err)
			return
		}

		fmt.Printf("HTML file %s written successfully\n", fileName)

	*/

	/*
		//i := interpreter("what is the best way to make a pizza?")
		//log.Println(i)
		//x := getResponse(i)
		//log.Println(x)

		instruct := `Provide a message for encouragement to create a positive state of mind or feeling of better well-being.`
		plugin := `**&timer: every other day at 9:20`

		p := instruct + plugin

		var f = func(s string) string {

			// this should be in a looped map
			if filterString(s, "**&timer:") {
				// run the plugin

			}

			return s
		}

		// returns the string without the plugin key
		var k = func(s string) string {

			if String(s, "**&") {
				// remove the plugin key from the string and what ever text is joining it
			}

			return s
		}

		runPlugins := f(p)
		res := runPolicies(k(p), bool(true))

		// handling the preprogrammed plugins
		// todo: map should be used to store all the plugin keys such as **&timer:

		_ = res

		log.Println(res)

		// sending sms
		//sendSMS(os.Getenv("TWILIO_PHONE_FROM"), os.Getenv("TWILIO_PHONE_TO"), quote)


	*/

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
	request := openai.ChatCompletionRequest{
		Model: inp.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: inp.prompt,
			},
		},

		MaxTokens:       inp.maxTokens,
		Temperature:     inp.temperature,
		TopP:            1,
		PresencePenalty: 0.6,
		Stop:            []string{"agent:", "person:"},
		//Stop:            []string{"tanaka:", "enquirer:", "reflector:", "prioritiser:", "planner:", "lister:", "decider:", "policy-decider:", "criticiser:", "recaller:", "tokensniffer:", "host:"},
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
