
func main() {
	//http.HandleFunc("/input", handleInput)
	//log.Fatal(http.ListenAndServe(":8080", addCorsHeaders(http.DefaultServeMux)))
	// Get the OpenAI API key from the .env file
	if err := godotenv.Load(); err != nil {
		log.Println("error loading .env file:", err)
		return
	}

	//var data map[string]interface{} = make(map[string]interface{})

	events := make([]interface{}, 0)

	for {

		var eventTime time.Time

		


		events = append(events, true)

		if time.Now().Hour() == 15 {
			// the prediction of dinner is done here
			data["tasks"][0]["task"] = "dinner"
			v := `can you suggest dinner for this evening?.` + taskContext

			x := getAnswer(v)

		}
	}

	s := make([]string, 0)

	// need to figure out some trigger which uses time to trigger the next question

	s = append(s, "can you suggest dinner for this evening?")
	//data["todo"] = s

	for _, v := range s {
		var done chan bool = make(chan bool)
		go func(v string, done chan bool) {
			x := getAnswer(v)
			log.Println(x)

			done <- true
		}(v, done)
		<-done
	}

}