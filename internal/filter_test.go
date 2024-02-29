package reporter

import "testing"

func TestFilterRule(t *testing.T) {
	priceFrom := 100.0
	priceTo := 200.0
	roomsFrom := 2
	roomsTo := 3
	areaFrom := 50.0
	areaTo := 100.0
	floorFrom := 1
	floorTo := 3
	isNotTopFloor := true

	filtersConf := Filters{
		Price: &RangeFilter[float64]{
			From: &priceFrom,
			To:   &priceTo,
		},
		Rooms: &RangeFilter[int]{
			From: &roomsFrom,
			To:   &roomsTo,
		},
		Area: &RangeFilter[float64]{
			From: &areaFrom,
			To:   &areaTo,
		},
		Floor: &RangeFilter[int]{
			From: &floorFrom,
			To:   &floorTo,
		},
		IsNotTopFloor: &isNotTopFloor,
	}

	listings := []Listing{
		{
			Id:         "1",
			Price:      150.0,
			Rooms:      2,
			Area:       60.0,
			Floor:      2,
			IsTopFloor: false,
		},
		{
			Id:         "2",
			Price:      250.0,
			Rooms:      3,
			Area:       70.0,
			Floor:      4,
			IsTopFloor: false,
		},
		{
			Id:         "3",
			Price:      150.0,
			Rooms:      2,
			Area:       60.0,
			Floor:      2,
			IsTopFloor: true,
		},
	}

	filtered := FilterRule(listings, filtersConf)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 listing, got %d", len(filtered))
	}
}
