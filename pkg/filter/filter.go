package filter

import (
	retrievalrules "github.com/niklc/listing-reporter/pkg/retrieval_rules"
	"github.com/niklc/listing-reporter/pkg/scraper"
)

func FilterCutoff(listings []scraper.Listing, cutoff []string) []scraper.Listing {
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

func GetNewCutoffs(listings []scraper.Listing) []string {
	ids := []string{}
	for i := 0; i < min(3, len(listings)); i++ {
		ids = append(ids, listings[i].Id)
	}
	return ids
}

func FilterRule(listings []scraper.Listing, filtersConf retrievalrules.Filters) []scraper.Listing {
	fil := []func(scraper.Listing, retrievalrules.Filters) bool{
		filterPrice,
		filterRooms,
		filterArea,
		filterFloor,
		filterIsNotTopFloor,
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

func filterRooms(listing scraper.Listing, filters retrievalrules.Filters) bool {
	return filterRange(listing.Rooms, filters.Rooms)
}

func filterArea(listing scraper.Listing, filters retrievalrules.Filters) bool {
	return filterRange(listing.Area, filters.Area)
}

func filterFloor(listing scraper.Listing, filters retrievalrules.Filters) bool {
	return filterRange(listing.Floor, filters.Floor)
}

func filterIsNotTopFloor(listing scraper.Listing, filters retrievalrules.Filters) bool {
	return filterBool(!listing.IsTopFloor, filters.IsNotTopFloor)
}

func filterRange[T int | float64](value T, rangeFilter *retrievalrules.RangeFilter[T]) bool {
	if rangeFilter == nil {
		return false
	}
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
