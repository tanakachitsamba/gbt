Here's a simple implementation for reading the txt file and converting its content to a string in Golang:

```golang
package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	// Replace "file.txt" with the path of your txt file
	content, err := ioutil.ReadFile("file.txt")
	if err != nil {
		log.Fatal(err)
	}

	text := string(content)
	fmt.Println(text)
}
```

In this code snippet, we are using the `ioutil.ReadFile()` function to read the contents of the file. It returns a byte slice containing the file content, which can be further converted to a string using `string(content)`. If there is an error while reading the file, it will log the error and exit the program.