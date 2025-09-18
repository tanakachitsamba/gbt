//go:build ignore

//nolint:all // Experimental utilities kept for reference.
package main

// global variable of the corpus need to uploaded to a db

// todo: to use a doc to extract
// todo: every 4000 charectors need to added a slice along with a struct of attributes of the document

func getCountOfChars(doc1 string) {
	countChars(doc1)
}

func countChars(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}
