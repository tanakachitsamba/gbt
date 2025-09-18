package main

import (
	"encoding/json"
	"log"
	"net/http"
)

var getClientFunc = getClient

type Request struct {
	Query string `json:"query"`
}

type Response struct {
	Result string `json:"result"`
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers before writing response
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Only allow POST requests
	//if r.Method != http.MethodPost {
	//	w.WriteHeader(http.StatusMethodNotAllowed)
	//	return
	//}

	// Parse the JSON payload
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	inp := Input{client: getClientFunc(), prompt: req.Query + "\n", model: "gpt-3.5-turbo-0613", temperature: 0.7, maxTokens: 250, systemMessage: `You are a gardening assistant. You provide concise and thoughtful answers to gardening topics.`}

	res, err := inp.getChatStreamResponse()
	if err != nil {
		log.Println("failed to obtain chat completion:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := Response{Result: res}

	// Convert the response object to JSON
	respJSON, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the content type and send the response
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respJSON); err != nil {
		log.Println("failed to write response:", err)
	}
}
