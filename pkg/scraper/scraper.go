package scraper

import (
	"io"
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
	Street string
	Rooms  string
	Area   string
	Floor  string
	Series string
	Price  string
}

func Scrape() ([]Listing, error) {
	body, err := fetch("lv/real-estate/flats/riga/centre/")
	if err != nil {
		return nil, err
	}

	listings, err := parse(body)
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

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parse(b string) ([]Listing, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(b))
	if err != nil {
		return []Listing{}, err
	}
	rows := doc.Find("[id^=tr_]")

	listings := []Listing{}
	rows.Each(func(_ int, row *goquery.Selection) {
		listings = append(listings, Listing{
			Id:     getId(row),
			Url:    getHref(row),
			Title:  getTextAt(row, 2),
			Img:    getImageSrc(row),
			Street: getTextAt(row, 3),
			Rooms:  getTextAt(row, 4),
			Area:   getTextAt(row, 5),
			Floor:  getTextAt(row, 6),
			Series: getTextAt(row, 7),
			Price:  getTextAt(row, 9),
		})
	})

	return listings, nil
}

func getId(row *goquery.Selection) string {
	val, _ := row.Attr("id")
	parts := strings.Split(val, "_")
	if len(parts) != 2 {
		return ""
	}

	return parts[1]
}

func getImageSrc(row *goquery.Selection) string {
	val, _ := row.Find("img").Attr("src")
	return val
}

func getHref(row *goquery.Selection) string {
	val, _ := row.Find("[href]").Attr("href")
	return val
}

func getTextAt(row *goquery.Selection, idx int) string {
	node := row.Children().Eq(idx).Text()
	return strings.TrimSpace(node)
}
