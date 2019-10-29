package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func main() {
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

	resp, err := http.Post(requestURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(body))
}
