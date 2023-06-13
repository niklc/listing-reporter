package main

import (
	"log"

	"github.com/niklc/listing-reporter/pkg/scraper"
)

func main() {
	res, err := scraper.Scrape()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(res)
}
