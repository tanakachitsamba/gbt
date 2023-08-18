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
	res                          string // maybe this could be a generic so that it can be both a slice, string or a null
}

type Plugin struct {
}

func main() {

	/*

			r := mux.NewRouter()
		r.HandleFunc("/", handleRequest).Methods("POST")

		// Enable CORS
		corsHandler := cors.Default().Handler(r)

		log.Println("Server listening on port 8080...")
		log.Fatal(http.ListenAndServe(":8080", corsHandler))

	*/

	inp := Input{client: getClient(), prompt: "Create questions for a job interview for a financial accounts manager in the uk" + "\n", model: "gpt-3.5-turbo-0613", temperature: 0.8, maxTokens: 1000, systemMessage: `Don't make assumptions about what values to plug into functions. Ask for clarification if a user request is ambiguous`}

	_, err, thing := inp.getChatStreamResponse()
	_ = thing

	if err != nil {
		log.Println(err, "error from the model")
	}

	//log.Println("answer: " + thing)

	// this creates files
	/*
		// check if output exist and if it does it gets moved to history under a unique name
		err = ioutil.WriteFile("output.txt", []byte(str), 0644)
		if err != nil {
			panic(err)
		}

	*/

	/*


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
