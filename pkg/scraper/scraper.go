package scraper

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const baseURI = "https://www.ss.lv/"

type Listing struct {
	Id     string
	Url    string
	Title  string
	Img    string
	Price  string
	Street string
	Rooms  int
	Area   int
	Floor  int
	Floors int
}

func Scrape() ([]Listing, error) {
	body, err := fetch("lv/real-estate/flats/riga/centre/")
	if err != nil {
		return nil, err
	}

	listings, err := parseBody(body)
	if err != nil {
		return nil, err
	}

	return listings, nil
}

func fetch(path string) (string, error) {
	res, err := http.Get(baseURI + path)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parseBody(b string) ([]Listing, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("[id^=tr_]").Each(func(i int, s *goquery.Selection) {
		html, _ := s.Html()
		fmt.Println(html)

		// s.Children().
	})

	return []Listing{}, nil
}
