package weather

import "fmt"

// Types for querying the city name

type SearchSuggestions struct {
	Suggestions []SearchSuggestion `json:"suggestions"`
}

type SearchSuggestion struct {
	Result string `json:"text"`
}

// Types for querying the lat/long

type LocationResults struct {
	Locations []Location `json:"locations"`
}

type Location struct {
	Name   string         `json:"name"`
	Extent LocationExtent `json:"extent"`
	// This also includes
	// {"feature": {"geometry": {"x": int, "y", int}}, "attributes": {"Score": int, "Addr_Type": "Postal"}}
}

type LocationExtent struct {
	Xmin float32 `json:"xmin"`
	Ymin float32 `json:"ymin"`
	Xmax float32 `json:"xmax"`
	Ymax float32 `json:"ymax"`
}

type LatLong struct {
	Lat  float32
	Long float32
}

// Weather
type Weather struct {
	WeatherTimes []WeatherTime
	Advisories   []Advisory
}

type Advisory struct {
	Description string
}

type WeatherTime struct {
	Label     string
	Temp      string // TODO: This should by a float, and keep track of °C/F
	ShortDesc string
	LongDesc  string
	// TODO: Air quality?
}

func (w WeatherTime) String() string {
	return fmt.Sprintf("[%s]: %s", w.Label, w.LongDesc)
}
