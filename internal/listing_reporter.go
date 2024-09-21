package reporter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws/session"
)

func Run() {
	awsSess, err := session.NewSession()
	if err != nil {
		log.Fatal("aws session creation failed: ", err)
	}

	rulesStore := NewRulesStore(awsSess)

	emailClient, rules, err := getEmailClientAndRules(awsSess, rulesStore)
	if err != nil {
		log.Fatal(err)
	}

	rulesSitesContent, err := fetchAllRulesSites(rules)
	if err != nil {
		log.Fatal(err)
	}

	rulesToUpdate := []RetrievalRule{}
	emails := []Email{}

	for _, rule := range rules {
		log.Println("processing config: " + rule.Name)

		printRule("config", rule)

		siteContent, ok := rulesSitesContent[rule.Name]
		if !ok {
			log.Fatalln("site contents not found for rule: ", rule.Name)
		}

		listings, err := Parse(siteContent)
		if err != nil {
			log.Println("site parse failed: ", err)
		}
		printListings("unfiltered", listings)

		newCutoffs := GetNewCutoffs(listings)
		log.Println("new cutoffs: " + strings.Join(newCutoffs, ", "))

		listings = FilterCutoff(listings, rule.Cutoffs)
		printListings("cutoff filtered", listings)

		listings = FilterRule(listings, rule.Filters)
		printListings("rules filtered", listings)

		if len(rule.Cutoffs) > 0 {
			for _, listing := range listings {
				emails = append(emails, Email{To: rule.Email, Listing: listing})
			}
			log.Printf("sending %d emails\n", len(listings))
		}

		if !isCutoffsEqual(rule.Cutoffs, newCutoffs) {
			rule.Cutoffs = newCutoffs
			rulesToUpdate = append(rulesToUpdate, rule)
		}
	}

	err = updateRulesSendEmails(rulesStore, emailClient, rulesToUpdate, emails)
	if err != nil {
		log.Fatal(err)
	}
}

func getEmailClientAndRules(awsSess *session.Session, rulesStore *RulesStore) (*EmailClient, []RetrievalRule, error) {
	type emailResult struct {
		client *EmailClient
		err    error
	}

	type rulesResult struct {
		err   error
		rules []RetrievalRule
	}

	emailChan := make(chan emailResult)
	rulesChan := make(chan rulesResult)

	go func() {
		credentialsBucket := NewCredentialsBucket(awsSess)
		emailConfig, emailToken, err := getEmailClientFiles(credentialsBucket)
		if err != nil {
			emailChan <- emailResult{err: err, client: nil}
			return
		}
		client, err := NewEmailClient(emailConfig, emailToken)
		if err != nil {
			err = fmt.Errorf("email client creation failed: %w", err)
		}
		emailChan <- emailResult{err: err, client: client}
	}()

	go func() {
		rules, err := rulesStore.Get()
		rulesChan <- rulesResult{err: err, rules: rules}
	}()

	emailRes := <-emailChan
	if emailRes.err != nil {
		return nil, nil, fmt.Errorf("config retrieval for email failed: %w", emailRes.err)
	}

	rulesRes := <-rulesChan
	if rulesRes.err != nil {
		return nil, nil, fmt.Errorf("token retrieval for email failed: %w", rulesRes.err)
	}

	return emailRes.client, rulesRes.rules, nil
}

func getEmailClientFiles(bucket *CredentialsBucket) ([]byte, []byte, error) {
	type result struct {
		err  error
		file []byte
	}

	configChan := make(chan result)
	tokenChan := make(chan result)

	get := func(name string, target chan result) {
		content, err := bucket.Get(name)
		target <- result{err: err, file: content}
	}

	go get("credentials.json", configChan)
	go get("token.json", tokenChan)

	configRes := <-configChan
	if configRes.err != nil {
		return nil, nil, fmt.Errorf("config retrieval for email failed: %w", configRes.err)
	}

	tokenRes := <-tokenChan
	if tokenRes.err != nil {
		return nil, nil, fmt.Errorf("token retrieval for email failed: %w", tokenRes.err)
	}

	return configRes.file, tokenRes.file, nil
}

func fetchAllRulesSites(rules []RetrievalRule) (map[string]string, error) {
	urls := map[string][]string{}
	for _, rule := range rules {
		_, ok := urls[rule.Url]
		if !ok {
			urls[rule.Url] = []string{}
		}
		urls[rule.Url] = append(urls[rule.Url], rule.Name)
	}

	type siteResult struct {
		err     error
		url     string
		content string
	}

	urlsLen := len(urls)

	sitesChan := make(chan siteResult, urlsLen)

	for url := range urls {
		go func(url string) {
			content, err := Fetch(url)
			sitesChan <- siteResult{err: err, url: url, content: content}
		}(url)
	}

	out := map[string]string{}

	for i := 0; i < urlsLen; i++ {
		res := <-sitesChan
		if res.err != nil {
			return nil, res.err
		}
		for _, name := range urls[res.url] {
			out[name] = res.content
		}
	}

	return out, nil
}

func printRule(name string, rule RetrievalRule) {
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

func printListings(name string, listings []Listing) {
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

func isCutoffsEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func updateRulesSendEmails(rulesStore *RulesStore, emailClient *EmailClient, rules []RetrievalRule, emails []Email) error {
	emailsLen := len(emails)

	rulesChan := make(chan error)
	emailsChan := make(chan error, emailsLen)

	go func() {
		rulesChan <- rulesStore.PutAll(rules)
	}()

	for _, email := range emails {
		go func() {
			emailsChan <- emailClient.SendListing(email)
		}()
	}

	rulesErr := <-rulesChan
	if rulesErr != nil {
		return fmt.Errorf("failed to put rules: %w ", rulesErr)
	}

	for i := 0; i < emailsLen; i++ {
		emailErr := <-emailsChan
		if emailErr != nil {
			return fmt.Errorf("failed sending listing email: %w", emailErr)
		}
	}

	return nil
}
