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

	// Process the query (you can implement your logic here)
	result := processQuery(req.Query)

	// Create the response object
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

func processQuery(query string) string {
	// Implement your query processing logic here

	client := getClient()

	//prompt := "create a html landing page about offering an accounting service to nurses, use tailwind for the css and make the website professional in design standards. only respond with the code\n"

	/*
		enc, err := encode(query + "\n")
		if err != nil {
			log.Fatalf("Encoding failed: %v", err)
		}
		lenOfInputTokens := enc.Count
		_ = lenOfInputTokens





			prompt, err := ioutil.ReadFile("prompt.txt")
			if err != nil {
				log.Fatalf("Error reading file: %v", err)
				os.Exit(1)
			}


	*/

	// after use old prompt gets stored in history using unique name

	//"gpt-3.5-turbo"

	inp := Input{client: client, prompt: query + "\n", model: "gpt-4", temperature: 0.9, maxTokens: 1500}

	var (
		str string
		err error
	)
	str, err = inp.getChatStreamResponse()
	_ = err
	_ = inp

	return str
}
