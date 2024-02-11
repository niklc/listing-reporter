package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	retrievalrules "github.com/niklc/listing-reporter/pkg/retrieval_rules"
	"github.com/niklc/listing-reporter/pkg/scraper"
)

func main() {
	rulesStore := retrievalrules.NewRulesStore()

	rules, err := rulesStore.Get()
	if err != nil {
		log.Fatal("config retrieval failed: ", err)
	}

	for _, rule := range rules {
		log.Println("processing config: " + rule.Name)

		printRule("config", rule)

		listings, err := scraper.Scrape(rule.Url)
		if err != nil {
			log.Println("listing retrieval failed: ", err)
		}
		printListings("unfiltered", listings)

		newCutoffs := getNewCutoffs(listings)
		log.Println("new cutoffs: " + strings.Join(newCutoffs, ", "))

		listings = filterCutoff(listings, rule.Cutoffs)
		printListings("cutoff filtered", listings)

		listings = filterRule(listings, rule.Filters)
		printListings("rules filtered", listings)

		// if len(rule.Cutoffs) > 0 {
		// sendListingEmail("")
		// log.Printf("sent %d emails\n", len(listings))
		// }

		rule.Cutoffs = newCutoffs
		err = rulesStore.Put(rule)
		if err != nil {
			log.Fatal("rule update failed: ", err)
		}
	}
}

func printRule(name string, rule retrievalrules.RetrievalRule) {
	headers := []string{"name", "email", "url", "filters", "cutoff"}
	filters, err := json.Marshal(rule.Filters)
	if err != nil {
		log.Fatal("print rule failed: ", err)
	}
	rows := [][]string{
		{
			rule.Name,
			rule.Email,
			rule.Url,
			string(filters),
			strings.Join(rule.Cutoffs, ","),
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
	s := map[string]bool{}
	for _, c := range cutoff {
		s[c] = true
	}
	for i, listing := range listings {
		for c := range s {
			if listing.Id == c {
				firstMatch = min(firstMatch, i)
				delete(s, c)
			}
		}
	}
	return listings[:firstMatch]
}

func getNewCutoffs(listings []scraper.Listing) []string {
	ids := []string{}
	for i := 0; i < min(3, len(listings)); i++ {
		ids = append(ids, listings[i].Id)
	}
	return ids
}

func filterRule(listings []scraper.Listing, filtersConf retrievalrules.Filters) []scraper.Listing {
	fil := []func(scraper.Listing, retrievalrules.Filters) bool{
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

func filterPrice(listing scraper.Listing, filters retrievalrules.Filters) bool {
	return filterRange(listing.Price, filters.Price)
}

func filterFloor(listing scraper.Listing, filters retrievalrules.Filters) bool {
	return filterRange(listing.Floor, filters.Floor)
}

func filterIsLastFloor(listing scraper.Listing, filters retrievalrules.Filters) bool {
	return filterBool(listing.IsTopFloor, filters.IsLastFloor)
}

func filterRange[T int | float64](value T, rangeFilter *retrievalrules.RangeFilter[T]) bool {
	if rangeFilter.From != nil && value < *rangeFilter.From {
		return true
	}
	if rangeFilter.To != nil && value > *rangeFilter.To {
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

func sendListingEmail(to string) {
	ctx := context.Background()

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	b, err = os.ReadFile("token.json")
	if err != nil {
		log.Fatalf("Unable to read token file: %v", err)
	}

	tok := &oauth2.Token{}
	err = json.Unmarshal(b, tok)
	if err != nil {
		log.Fatalf("Unable to parse token file: %v", err)
	}

	client := config.Client(ctx, tok)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	subject := "test"

	_, err = srv.Users.Messages.Send("me", &gmail.Message{
		Raw: base64.StdEncoding.EncodeToString([]byte(
			"From: 'me'\r\n" +
				"To:  " + to + "\r\n" +
				"Subject: " + subject + "\r\n" +
				"Content-Type: text/html; charset=UTF-8\r\n" +
				"Content-Transfer-Encoding: 8bit\r\n" +
				"\r\n" +
				"<table border=\"1\" cellpadding=\"10\" cellspacing=\"0\">" +
				"<tr><td>test</td><td>test</td></tr>" +
				"</table>",
		)),
	}).Do()
	if err != nil {
		log.Fatal(err)
	}
}
