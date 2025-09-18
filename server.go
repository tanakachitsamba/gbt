package main

import (
	"encoding/json"
	"net/http"

	appopenai "guava/pkg/openai"
)

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

	// Parse the JSON payload
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	config := appopenai.ResponseConfig{
		Model:           appopenai.ModelGPT4oMini,
		Temperature:     appopenai.Float32Ptr(0.7),
		MaxOutputTokens: appopenai.IntPtr(250),
		Instructions:    `You are a gardening assistant. You provide concise and thoughtful answers to gardening topics.`,
	}

	inp := Input{
		client: getClient(),
		prompt: req.Query + "\n",
		config: config,
	}

	result, err := inp.getChatStreamResponse()
	if err != nil {
		http.Error(w, "failed to generate response", http.StatusInternalServerError)
		return
	}

	resp := Response{Result: result}

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
