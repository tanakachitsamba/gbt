//go:build ignore
// +build ignore

//nolint:all // Legacy scraping prototype not used in production.
package main

import (
	"log"

	"github.com/gocolly/colly"
)

func scrape(site string) string {
	c := colly.NewCollector()
	var res string
	// Callback function that will be called when a visited HTML element is found
	c.OnHTML("body", func(e *colly.HTMLElement) {
		// Print the text of the body element
		res = e.Text
	})

	// Error handling callback
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// Start scraping
	err := c.Visit(site)
	if err != nil {
		log.Fatal(err)
	}

	return res
}
