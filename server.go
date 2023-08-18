package main

import (
	"encoding/json"
	"net/http"
)

type Request struct {
	Query string `json:"query"`
}

type Response struct {
	Result string `json:"result"`
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON payload
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Process the query
	// Create the response object
	//dataAndTime := time.DateTime

	//location := "Nottingham, UK"
	//manifest := `Don't answer questions that require locations or date/time like this "I'm sorry, as a software development agent, I don't have access to real-time weather and agricultural information.", use the data and time provided and the location provided to answer` + "the date and time is: " + dataAndTime + " and the location is: " + location + `\n`
	inp := Input{client: getClient(), prompt: req.Query + "\n", model: "gpt-3.5-turbo-0613", temperature: 0.8, maxTokens: 1000, systemMessage: `You are a gardening assistant.Format the response so that it is readable.`}

	res, err := inp.getChatStreamResponse()
	_ = err

	resp := Response{Result: res}

	// Convert the response object to JSON
	respJSON, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the content type and send the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(respJSON)
}

func getInput(prompt string) Input {
	//inp := Input{client: getClient(), prompt: prompt + "\n", model: "gpt-3.5-turbo", temperature: 0.8, maxTokens: 400, systemMessage: `You are a gardening helper, you do not respond with any response that starts with "As an AI model".  You will only answer questions about gardening, allotmenting and any related topics to do with plant growing. If you recieve a question thats not related to gardening or allotments, you will politely decline to answer and respond with "I cannot answer that this question because it's not gardening related".  Use a friendly tone`}

	inp := Input{client: getClient(), prompt: prompt + "\n", model: "gpt-4", temperature: 0.8, maxTokens: 1000, systemMessage: `You are a software development agent, your job is to write high quality non-buggy code`}

	return inp
}

func processQuery(query string) string {
	var (
		res string
		err error
		inp Input = getInput(query)
	)
	res, err = inp.getChatStreamResponse()
	_ = err
	_ = inp

	return res
}
