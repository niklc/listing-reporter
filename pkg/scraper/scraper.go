package scraper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const baseUrl = "https://www.ss.lv"

type Listing struct {
	Id         string
	Url        string
	Title      string
	Img        string
	Street     string
	Rooms      int
	Area       float64
	Floor      int
	Floors     int
	IsTopFloor bool
	Series     string
	Price      float64
}

func Scrape(url string) ([]Listing, error) {
	body, err := fetch(url)
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
	res, err := http.Get(baseUrl + path)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

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
	if rows.Length() == 0 {
		return []Listing{}, fmt.Errorf("no rows found")
	}

	listings := []Listing{}
	rows.Each(func(_ int, row *goquery.Selection) {
		if isBannerRow(row) {
			return
		}

		id, err := getId(row)
		if err != nil {
			log.Println(err)
			return
		}
		url, err := getHref(row)
		if err != nil {
			log.Printf("row %s: %s", id, err)
			return
		}
		title, err := getTextAt(row, 2)
		if err != nil {
			log.Printf("row %s: %s", id, err)
			return
		}
		img, err := getImageSrc(row)
		if err != nil {
			log.Printf("row %s: %s", id, err)
			return
		}
		street, err := getTextAt(row, 3)
		if err != nil {
			log.Printf("row %s: %s", id, err)
			return
		}
		rooms, err := getIntAt(row, 4)
		if err != nil {
			log.Printf("row %s: %s", id, err)
			return
		}
		area, err := getFloatAt(row, 5)
		if err != nil {
			log.Printf("row %s: %s", id, err)
			return
		}
		floor, floors, err := getFloorAndFloorsAt(row, 6)
		if err != nil {
			log.Printf("row %s: %s", id, err)
			return
		}
		series, err := getTextAt(row, 7)
		if err != nil {
			log.Printf("row %s: %s", id, err)
			return
		}
		price, err := getPriceAt(row, 9)
		if err != nil {
			log.Printf("row %s: %s", id, err)
			return
		}
		listings = append(listings, Listing{
			Id:         id,
			Url:        baseUrl+url,
			Title:      title,
			Img:        img,
			Street:     street,
			Rooms:      rooms,
			Area:       area,
			Floor:      floor,
			Floors:     floors,
			IsTopFloor: floor == floors,
			Series:     series,
			Price:      price,
		})
	})

	return listings, nil
}

func isBannerRow(row *goquery.Selection) bool {
	val, _ := row.Attr("id")
	return strings.Contains(val, "bnr")
}

func getId(row *goquery.Selection) (string, error) {
	val, _ := row.Attr("id")
	if val == "" {
		return "", fmt.Errorf("no id found")
	}
	parts := strings.Split(val, "_")
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected id format: %s", val)
	}
	return parts[1], nil
}

func getImageSrc(row *goquery.Selection) (string, error) {
	val, _ := row.Find("img").Attr("src")
	if val == "" {
		return "", fmt.Errorf("no image found")
	}
	return val, nil
}

func getHref(row *goquery.Selection) (string, error) {
	val, _ := row.Find("[href]").Attr("href")
	if val == "" {
		return "", fmt.Errorf("no href found")
	}
	return val, nil
}

func getPriceAt(row *goquery.Selection, idx int) (float64, error) {
	str, err := getTextAt(row, idx)
	if err != nil {
		return 0, err
	}
	r, err := regexp.Compile(`[0-9,\.]+`)
	if err != nil {
		return 0, err
	}
	part := r.FindString(str)
	if part == "" {
		return 0, fmt.Errorf("unexpected price format: %s", str)
	}
	part = strings.ReplaceAll(part, ",", "")
	price, err := strconv.ParseFloat(part, 64)
	if err != nil {
		return 0, err
	}
	return price, nil
}

func getFloorAndFloorsAt(row *goquery.Selection, idx int) (int, int, error) {
	str, err := getTextAt(row, idx)
	if err != nil {
		return 0, 0, err
	}
	r, err := regexp.Compile(`[0-9]+`)
	if err != nil {
		return 0, 0, err
	}
	parts := r.FindAllString(str, -1)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected floor format: %s", str)
	}
	floor, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	floors, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return floor, floors, nil
}

func getIntAt(row *goquery.Selection, idx int) (int, error) {
	str, err := getTextAt(row, idx)
	if err != nil {
		return 0, err
	}
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func getFloatAt(row *goquery.Selection, idx int) (float64, error) {
	str, err := getTextAt(row, idx)
	if err != nil {
		return 0, err
	}
	float, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, err
	}
	return float, nil
}

func getTextAt(row *goquery.Selection, idx int) (string, error) {
	node := row.Children().Eq(idx).Text()
	if node == "" {
		return "", fmt.Errorf("no text found")
	}
	return strings.TrimSpace(node), nil
}
