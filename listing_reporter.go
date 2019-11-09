package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
)

func main() {
	body := fetch()

	// bodyBytes, err := ioutil.ReadFile("response_body.html")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// body := string(bodyBytes)

	listings := parse(body)

	output(listings)
}

func fetch() string {
	requestURL := "https://www.ss.com/lv/transport/cars/bmw/3-series/sell/filter/"

	data := url.Values{}
	data.Set("topt[8][min]", "3000")
	data.Set("topt[8][max]", "6000")
	data.Set("topt[18][min]", "2001")
	data.Set("topt[18][max]", "2008")
	data.Set("topt[15][min]", "2.0")
	data.Set("opt[34]", "494")
	data.Set("opt[35]", "496")
	data.Set("opt[32]", "484")

	resp, err := http.Post(
		requestURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)

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

	return parsedListings
}

func output(listings []map[string]string) {
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
