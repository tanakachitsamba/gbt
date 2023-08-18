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

	inp := Input{client: getClient(), prompt: req.Query + "\n", model: "gpt-3.5-turbo-0613", temperature: 0.7, maxTokens: 250, systemMessage: `You are a gardening assistant. You provide concise and thoughtful answers to gardening topics.`}

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
