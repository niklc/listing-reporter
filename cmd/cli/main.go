package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/niklc/listing-reporter/pkg/scraper"
)

func main() {
	listings, err := scraper.Scrape()
	if err != nil {
		log.Fatal(err)
	}

	writter := tabwriter.NewWriter(
		os.Stdout,
		0,
		0,
		2,
		' ',
		0,
	)
	header := []string{"id", "url", "title", "price"}
	for _, h := range header {
		fmt.Fprintf(writter, "%s\t", h)
	}
	fmt.Fprintln(writter)
	for _, listing := range listings {
		cleanedTitle := strings.ReplaceAll(listing.Title, "\n", " ")
		row := []string{
			listing.Id,
			listing.Url,
			cleanedTitle,
			strconv.FormatFloat(listing.Price, 'f', -1, 64),
		}
		for _, r := range row {
			fmt.Fprintf(writter, "%s\t", r)
		}
		fmt.Fprintln(writter)
	}

	writter.Flush()
}
