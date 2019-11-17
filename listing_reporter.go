package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const filtersPath = "filters.json"
const baseURI = "https://www.ss.lv"

type filter struct {
	CutoffID   string            `json:"cutoffID"`
	Parameters map[string]string `json:"parameters"`
	Path       string            `json:"path"`
	Subsection bool              `json:"subsection"`
}

var filters []filter

func main() {
	config()

	for {
		for filterIndex, filter := range filters {
			body := fetch(filter.Parameters, filter.Path)

			// bodyBytes, err := ioutil.ReadFile("response_body.html")
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// body := string(bodyBytes)

			listings := parse(body, filterIndex, filter.CutoffID, filter.Subsection)

			output(listings)
		}

		time.Sleep(time.Hour)
	}
}

func config() {
	data, err := ioutil.ReadFile(filtersPath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &filters)
	if err != nil {
		log.Fatal(err)
	}
}

func fetch(params map[string]string, path string) string {
	u, _ := url.ParseRequestURI(baseURI)

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	if err != nil {
		panic(err)
	}

	client := http.Client{Jar: jar}

	resp, err := client.Get(u.String())

	if err != nil {
		log.Fatal(err)
	}

	u.Path = path

	fParams := url.Values{}
	for key, value := range params {
		fParams.Add(key, value)
	}
	resp, err = client.PostForm(u.String(), fParams)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	return string(body)
}

func parse(body string, filterIndex int, cutoffID string, isSubsection bool) []map[string]string {
	replaceString := " "
	replacer := strings.NewReplacer(
		"\n", replaceString,
		"\r\n", replaceString,
	)
	bodyFormatted := replacer.Replace(body)

	regex := regexp.MustCompile(`<tr id="tr_\d+">(.+?)</tr>`)
	rows := regex.FindAllString(bodyFormatted, -1)

	type mapIndexPattern struct {
		index   int
		pattern string
	}

	basicPattern := `>(?:<b>)?(.+?)(?:</b>)?<`

	parseMap := map[string]mapIndexPattern{
		"id":          mapIndexPattern{1, `id="im(\d+)"`},
		"description": mapIndexPattern{2, `">(?:<b>)?(.+?)(?:</b>)?</a`},
		"url":         mapIndexPattern{1, `a href="(.+?)"`},
		"image":       mapIndexPattern{1, `img src="(.+?)"`},
	}

	if isSubsection {
		parseMap["model"] = mapIndexPattern{3, basicPattern}
		parseMap["year"] = mapIndexPattern{4, basicPattern}
		parseMap["volume"] = mapIndexPattern{5, basicPattern}
		parseMap["mileage"] = mapIndexPattern{6, basicPattern}
		parseMap["price"] = mapIndexPattern{7, basicPattern}
	} else {
		parseMap["year"] = mapIndexPattern{3, basicPattern}
		parseMap["volume"] = mapIndexPattern{4, basicPattern}
		parseMap["mileage"] = mapIndexPattern{5, basicPattern}
		parseMap["price"] = mapIndexPattern{6, basicPattern}
	}

	parsedListings := make([]map[string]string, 0)
	for _, row := range rows {

		regex = regexp.MustCompile(`<td(.+?)</td>`)
		cols := regex.FindAllString(row, -1)

		colsMap := make(map[string]string)

		for col, map1 := range parseMap {
			if map1.index >= len(cols) {
				colsMap[col] = "parse err"
				continue
			}
			regex = regexp.MustCompile(map1.pattern)
			matches := regex.FindStringSubmatch(cols[map1.index])
			if len(matches) > 1 {
				colsMap[col] = matches[1]
			}
		}

		parsedListings = append(parsedListings, colsMap)
	}

	filteredListings := filterForNewListings(parsedListings, filterIndex, cutoffID)

	return filteredListings
}

func filterForNewListings(listings []map[string]string, filterIndex int, cutoffID string) []map[string]string {

	newListings := []map[string]string{}

	for _, listing := range listings {

		if listing["id"] == cutoffID {
			break
		}

		newListings = append(newListings, listing)
	}

	if len(newListings) > 0 {
		updateCutoff(filterIndex, newListings[0]["id"])

		if len(listings) == len(newListings) {
			return newListings[0:1]
		}
	}

	return newListings
}

func updateCutoff(filterIndex int, cutoffID string) {
	filters[filterIndex].CutoffID = cutoffID

	json, _ := json.MarshalIndent(filters, "", "    ")

	err := ioutil.WriteFile(filtersPath, json, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func output(listings []map[string]string) {
	outputTerminal(listings)

	for _, listing := range listings {
		outputEmail(listing)
	}
}

func outputTerminal(listings []map[string]string) {
	writter := tabwriter.NewWriter(
		os.Stdout,
		0,
		0,
		2,
		' ',
		0,
	)
	keys := []string{"price", "mileage", "year", "description"}
	for _, listing := range listings {
		for _, key := range keys {
			fmt.Fprintf(writter, "%s: %.50s\t", key, listing[key])
		}
		fmt.Fprintf(writter, "\n")
	}
	writter.Flush()
}

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func outputEmail(listing map[string]string) {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	to := ""
	subject := "ss.lv listing " + listing["id"]

	content := []byte("From: 'me'\r\n" +
		"To:  " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: 8bit\r\n" +
		"\r\n" +
		"<table border=\"1\" cellpadding=\"10\" cellspacing=\"0\">")

	order := []string{"url", "image", "price", "model", "year", "volume", "mileage", "description"}

	for _, index := range order {
		value := listing[index]
		if listing[index] == "" {
			continue
		}
		switch index {
		case "url":
			value = "<a href=\"" + baseURI + value + "\">" + value + "</a>"
		case "image":
			value = "<img src=\"" + value + "\">"
		}
		content = append(content, []byte("<tr><td>"+index+"</td><td>"+value+"</td></tr>")...)
	}

	content = append(content, []byte("</table>")...)

	var message gmail.Message
	message.Raw = base64.StdEncoding.EncodeToString(content)
	message.Raw = strings.Replace(message.Raw, "/", "_", -1)
	message.Raw = strings.Replace(message.Raw, "+", "-", -1)
	message.Raw = strings.Replace(message.Raw, "=", "", -1)

	_, err = srv.Users.Messages.Send("me", &message).Do()
	if err != nil {
		log.Fatalf("Unable to send message: %v", err)
	}
}
