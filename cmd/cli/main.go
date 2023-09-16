package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/niklc/listing-reporter/pkg/scraper"
)

func main() {
	configs := getConfigs()
	for _, config := range configs {
		listings, err := scraper.Scrape(config.url)
		if err != nil {
			log.Fatal(err)
		}
		printConfig(config)
		printListings(listings)
	}
}

type config struct {
	name    string
	email   string
	url     string
	filters string
	cutoff  []string
}

func getConfigs() []config {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           "personal",
	}))

	svc := dynamodb.New(sess)

	tableName := "listing-reporter"
	res, err := svc.Scan(&dynamodb.ScanInput{TableName: &tableName})
	if err != nil {
		log.Fatal(err)
	}

	configs := []config{}
	for _, item := range res.Items {
		cutoff := *item["cutoff"].S
		configs = append(configs, config{
			name:    *item["name"].S,
			email:   *item["email"].S,
			url:     *item["url"].S,
			filters: *item["filters"].S,
			cutoff:  strings.Split(cutoff, ","),
		})
	}

	if len(configs) == 0 {
		log.Fatal("no configs found")
	}

	return configs
}

func printConfig(config config) {
	headers := []string{"name", "email", "url", "filters", "cutoff"}
	rows := [][]string{
		{
			config.name,
			config.email,
			config.url,
			config.filters,
			strings.Join(config.cutoff, ","),
		},
	}
	printCsv(headers, rows)
}

func printListings(listings []scraper.Listing) {
	header := []string{"id", "url", "title", "price"}
	rows := [][]string{}
	for _, listing := range listings {
		cleanedTitle := strings.ReplaceAll(listing.Title, "\n", " ")
		rows = append(rows,
			[]string{
				listing.Id,
				listing.Url,
				cleanedTitle,
				strconv.FormatFloat(listing.Price, 'f', -1, 64),
			},
		)
	}
	printCsv(header, rows)
}

func printCsv(headers []string, rows [][]string) {
	writer := tabwriter.NewWriter(
		os.Stdout,
		0,
		0,
		2,
		' ',
		0,
	)
	for _, h := range headers {
		fmt.Fprintf(writer, "%s\t", h)
	}
	fmt.Fprintln(writer)
	for _, row := range rows {
		for _, r := range row {
			fmt.Fprintf(writer, "%s\t", r)
		}
		fmt.Fprintln(writer)
	}
	writer.Flush()
}
