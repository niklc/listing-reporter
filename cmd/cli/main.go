package main

import (
	"encoding/json"
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
		printConfig("config", config)
		printListings("raw", listings)

		listings = filterCutoff(listings, config.cutoff)
		printListings("cutoff filtered", listings)

		newCutoffs := getNewCutoffs(listings)
		log.Println("cutoffs:" + strings.Join(newCutoffs, ", "))

		filters := getFilters(config.filters)

		listings = filterConfig(listings, filters)
		printListings("filtered", listings)
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

func printConfig(name string, config config) {
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
	printCsv(name, headers, rows)
}

func printListings(name string, listings []scraper.Listing) {
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
	printCsv(name, header, rows)
}

func printCsv(name string, headers []string, rows [][]string) {
	writer := tabwriter.NewWriter(
		os.Stdout,
		0,
		0,
		2,
		' ',
		0,
	)
	if name != "" {
		fmt.Fprintf(writer, "%s:\n", name)
	}
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

func filterCutoff(listings []scraper.Listing, cutoff []string) []scraper.Listing {
	firstMatch := len(listings)
	for i, listing := range listings {
		for _, c := range cutoff {
			if listing.Id == c {
				firstMatch = i - 1
				break
			}
		}
	}
	return listings[:firstMatch]
}

func getNewCutoffs(listings []scraper.Listing) []string {
	ids := []string{}
	for i := len(listings) - 1; i >= len(listings)-3; i-- {
		ids = append(ids, listings[i].Id)
	}
	return ids
}

type rangeFilter struct {
	From float64
	To   float64
}

type filters struct {
	Price       rangeFilter
	Floor       rangeFilter
	IsLastFloor *bool
}

func getFilters(filtersJson string) filters {
	filters := filters{}
	err := json.Unmarshal([]byte(filtersJson), &filters)
	if err != nil {
		log.Fatal(err)
	}
	return filters
}

func filterConfig(listings []scraper.Listing, filtersConf filters) []scraper.Listing {
	fil := []func(scraper.Listing, filters) bool{
		filterPrice,
		filterFloor,
		filterIsLastFloor,
	}
	remaining := []scraper.Listing{}
out:
	for _, listing := range listings {
		for _, f := range fil {
			if f(listing, filtersConf) {
				continue out
			}
		}
		remaining = append(remaining, listing)
	}
	return remaining
}

func filterPrice(listing scraper.Listing, filters filters) bool {
	return filterRange(listing.Price, filters.Price)
}

func filterFloor(listing scraper.Listing, filters filters) bool {
	return filterRange(float64(listing.Floor), filters.Floor)
}

func filterIsLastFloor(listing scraper.Listing, filters filters) bool {
	return filterBool(listing.IsTopFloor, filters.IsLastFloor)
}

func filterRange(value float64, rangeFilter rangeFilter) bool {
	if rangeFilter.From != 0 && value < rangeFilter.From {
		return true
	}
	if rangeFilter.To != 0 && value > rangeFilter.To {
		return true
	}
	return false
}

func filterBool(value bool, filter *bool) bool {
	if filter == nil {
		return false
	}
	return value != *filter
}
