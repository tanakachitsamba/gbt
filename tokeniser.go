package main

import (
	"encoding/json"
	"log"
	"os/exec"
)

// define a struct to hold the JSON output
type EncoderOutput struct {
	Tokens []int `json:"tokens"`
	Count  int   `json:"count"`
}

// encode function takes a string to be encoded and returns encoded tokens and their count
func encode(str string) (EncoderOutput, error) {
	// run the JavaScript file with Node.js
	cmd := exec.Command("node", "encoder.js", str)

	// get the output of the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Command failed with error: %v. Output: %s", err, output)
		return EncoderOutput{}, err
	}

	// parse the JSON output
	var result EncoderOutput
	err = json.Unmarshal(output, &result)
	if err != nil {
		log.Printf("Failed to parse JSON output: %v", err)
		return EncoderOutput{}, err
	}

	return result, nil
}
