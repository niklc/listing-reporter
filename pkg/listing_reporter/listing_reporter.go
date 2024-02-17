package listingreporter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/niklc/listing-reporter/pkg/email"
	"github.com/niklc/listing-reporter/pkg/filter"
	retrievalrules "github.com/niklc/listing-reporter/pkg/retrieval_rules"
	"github.com/niklc/listing-reporter/pkg/scraper"
)

func Run() {
	awsSess, err := session.NewSession()
	if err != nil {
		log.Fatal("aws session creation failed: ", err)
	}

	rulesStore := retrievalrules.NewRulesStore(awsSess)

	emailClient, err := email.NewEmailClient(awsSess)
	if err != nil {
		log.Fatal("email client creation failed: ", err)
	}

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

		newCutoffs := filter.GetNewCutoffs(listings)
		log.Println("new cutoffs: " + strings.Join(newCutoffs, ", "))

		listings = filter.FilterCutoff(listings, rule.Cutoffs)
		printListings("cutoff filtered", listings)

		listings = filter.FilterRule(listings, rule.Filters)
		printListings("rules filtered", listings)

		if len(rule.Cutoffs) > 0 {
			for _, listing := range listings {
				emailClient.SendListing(rule.Email, listing)
			}
			log.Printf("sent %d emails\n", len(listings))
		}

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
