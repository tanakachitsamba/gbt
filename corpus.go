package main

// global variable of the corpus need to uploaded to a db

// getCountOfChars returns the number of Unicode characters in the provided document.
func getCountOfChars(doc1 string) int {
	return countChars(doc1)
}

func countChars(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}
