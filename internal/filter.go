package reporter

func FilterCutoff(listings []Listing, cutoff []string) []Listing {
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

func GetNewCutoffs(listings []Listing) []string {
	ids := []string{}
	for i := 0; i < min(3, len(listings)); i++ {
		ids = append(ids, listings[i].Id)
	}
	return ids
}

func FilterRule(listings []Listing, filtersConf Filters) []Listing {
	fil := []func(Listing, Filters) bool{
		filterPrice,
		filterRooms,
		filterArea,
		filterFloor,
		filterIsNotTopFloor,
	}
	remaining := []Listing{}
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

func filterPrice(listing Listing, filters Filters) bool {
	return filterRange(listing.Price, filters.Price)
}

func filterRooms(listing Listing, filters Filters) bool {
	return filterRange(listing.Rooms, filters.Rooms)
}

func filterArea(listing Listing, filters Filters) bool {
	return filterRange(listing.Area, filters.Area)
}

func filterFloor(listing Listing, filters Filters) bool {
	return filterRange(listing.Floor, filters.Floor)
}

func filterIsNotTopFloor(listing Listing, filters Filters) bool {
	return filterBool(!listing.IsTopFloor, filters.IsNotTopFloor)
}

func filterRange[T int | float64](value T, rangeFilter *RangeFilter[T]) bool {
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
