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
	// Set CORS headers before writing response
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5173")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Only allow POST requests
	//if r.Method != http.MethodPost {
	//	w.WriteHeader(http.StatusMethodNotAllowed)
	//	return
	//}

	// Parse the JSON payload
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// "gpt-3.5-turbo-0613"
	// `You are a gardening assistant. You provide concise and thoughtful answers to gardening topics.`
	inp := Input{client: getClient(), prompt: req.Query + "\n", model: "gpt-4-0613", temperature: 0.7, maxTokens: 250, systemMessage: `You are a personal assistant.`}

	res, err := inp.getChatStreamResponse()
	_ = err

	// Convert the response object to JSON
	respJSON, err := json.Marshal(Response{Result: res})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the content type and send the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(respJSON)
}
