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

func main() {
	for {
		body := fetch()

		// bodyBytes, err := ioutil.ReadFile("response_body.html")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// body := string(bodyBytes)

		listings := parse(body)

		output(listings)

		time.Sleep(time.Hour)
	}
}

func fetch() string {
	u, _ := url.ParseRequestURI("https://www.ss.lv")

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	if err != nil {
		panic(err)
	}

	client := http.Client{Jar: jar}

	resp, err := client.Get(u.String())

	if err != nil {
		log.Fatal(err)
	}

	u.Path = "/lv/transport/cars/bmw/3-series/sell/filter/"

	resp, err = client.PostForm(u.String(), url.Values{
		"topt[8][min]":  {"3000"},
		"topt[8][max]":  {"6000"},
		"topt[18][min]": {"2001"},
		"topt[18][max]": {"2008"},
		"topt[15][min]": {"2.0"},
		"opt[32]":       {"484"},
		"opt[34]":       {"494"},
		"opt[35]":       {"496"},
	})

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

func parse(body string) []map[string]string {
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

	basicPattern := `>(.+?)<`

	parseMap := map[string]mapIndexPattern{
		"id":          mapIndexPattern{1, `id="im\d+"`},
		"description": mapIndexPattern{2, `">(.+?)</a`},
		"url":         mapIndexPattern{1, `a href="(.+?)"`},
		"image":       mapIndexPattern{1, `img src="(.+?)"`},
		"model":       mapIndexPattern{3, basicPattern},
		"year":        mapIndexPattern{4, basicPattern},
		"volume":      mapIndexPattern{5, basicPattern},
		"mileage":     mapIndexPattern{6, basicPattern},
		"price":       mapIndexPattern{7, basicPattern},
	}

	parsedListings := make([]map[string]string, 0)
	for _, row := range rows {

		regex = regexp.MustCompile(`<td(.+?)</td>`)
		cols := regex.FindAllString(row, -1)

		colsMap := make(map[string]string)

		for col, map1 := range parseMap {
			regex = regexp.MustCompile(map1.pattern)
			colsMap[col] = regex.FindString(cols[map1.index])
		}

		parsedListings = append(parsedListings, colsMap)
	}

	filteredListings := filterForNewListings(parsedListings)

	return filteredListings
}

func filterForNewListings(listings []map[string]string) []map[string]string {
	cutoffID := getCutoffID()

	newListings := []map[string]string{}

	for _, listing := range listings {

		if listing["id"] == cutoffID {
			break
		}

		newListings = append(newListings, listing)
	}

	if len(newListings) > 0 {
		setCutoffID(newListings[0]["id"])
	}

	if len(listings) == len(newListings) {
		return newListings[0:1]
	}

	return newListings
}

func getCutoffID() string {
	data, err := ioutil.ReadFile("filters.json")
	if err != nil {
		log.Fatal(err)
	}

	var filters []interface{}

	err = json.Unmarshal(data, &filters)
	if err != nil {
		log.Fatal(err)
	}
	filter := filters[0].(map[string]interface{})
	id := filter["cutoffID"].(string)

	return id
}

func setCutoffID(id string) {
	fileName := "filters.json"
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	var filters []interface{}

	err = json.Unmarshal(data, &filters)
	if err != nil {
		log.Fatal(err)
	}

	filter := filters[0].(map[string]interface{})
	filter["cutoffID"] = id

	filters[0] = filter

	res, _ := json.MarshalIndent(filters, "", "    ")

	err = ioutil.WriteFile(fileName, res, 0644)
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

	user := "me"

	to := ""
	subject := "ss.lv listing"

	var message gmail.Message
	temp := []byte("From: 'me'\r\n" +
		"To:  " + to + "\r\n" +
		"Subject: " + subject + " " + listing["id"] + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: 8bit\n\n" +
		"<br>" + listing["url"] +
		"<br><" + listing["image"] + ">" +
		"<br>" + listing["model"] +
		"<br>" + listing["year"] +
		"<br>" + listing["volume"] +
		"<br>" + listing["mileage"] +
		"<br>" + listing["price"] +
		"<br>" + listing["description"])

	message.Raw = base64.StdEncoding.EncodeToString(temp)
	message.Raw = strings.Replace(message.Raw, "/", "_", -1)
	message.Raw = strings.Replace(message.Raw, "+", "-", -1)
	message.Raw = strings.Replace(message.Raw, "=", "", -1)

	_, err = srv.Users.Messages.Send(user, &message).Do()
	if err != nil {
		log.Fatalf("Unable to send message: %v", err)
	}
}
