package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
)

func WriteHTMLFile(fileName string, content string) error {
	// Check if the file already exists
	if _, err := os.Stat(fileName); err == nil {
		// Generate a random three-digit number
		randomNum := generateRandomNumber()

		// Rename the existing file
		newFileName := fmt.Sprintf("%s.html", randomNum)
		err := os.Rename(fileName, newFileName)
		if err != nil {
			return fmt.Errorf("error renaming file: %w", err)
		}

		fmt.Printf("Renamed existing file from %s to %s\n", fileName, newFileName)
	}

	// Write the new file
	err := ioutil.WriteFile(fileName, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func generateRandomNumber() string {
	// Generate a random three-digit number
	randomNum := fmt.Sprintf("%02d", randomInt(0, 99))
	return randomNum
}

func randomInt(min, max int) int {
	// Return a random integer between min and max (inclusive)
	return min + rand.Intn(max-min+1)
}
