package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

// Exported types

// Latitude and longitude
type LatLong struct {
	Lat  float32
	Long float32
}

// Weather result
type Weather struct {
	WeatherTimes []WeatherTime
	Advisories   []Advisory
}

// Active advisories
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

type WeatherGetter interface {
	GetCity(string) (string, error)
	GetLatLong(string) (*LatLong, error)
	GetWeather(*LatLong) (*Weather, error)
}

func MakeWeatherGetter(userAgent string) WeatherGetter {
	return &getWeatherData{
		cityurl,
		latlongurl,
		weatherurl,
		userAgent,
		&http.Client{},
	}
}

// Unexported types

// Types for querying the city name

type searchSuggestions struct {
	Suggestions []searchSuggestion `json:"suggestions"`
}

type searchSuggestion struct {
	Result string `json:"text"`
}

// Types for querying the lat/long

type locationResults struct {
	Locations []location `json:"locations"`
}

type location struct {
	Name   string         `json:"name"`
	Extent locationExtent `json:"extent"`
	// This also includes
	// {"feature": {"geometry": {"x": int, "y", int}}, "attributes": {"Score": int, "Addr_Type": "Postal"}}
}

type locationExtent struct {
	Xmin float32 `json:"xmin"`
	Ymin float32 `json:"ymin"`
	Xmax float32 `json:"xmax"`
	Ymax float32 `json:"ymax"`
}

type getWeatherData struct {
	CityUrl    string       // Url to check for city location
	LatLongUrl string       // Url to check for lat/long location
	WeatherUrl string       // Url to check for the weather
	UserAgent  string       // User agent to use when sending requests
	Client     *http.Client // HTTP client to use when sending requests
}

// Get the correct city name from a query
func (w *getWeatherData) GetCity(query string) (string, error) {
	if w.Client == nil {
		w.Client = &http.Client{}
	}
	var queryUrl string
	if w.CityUrl != "" {
		queryUrl = w.CityUrl
	} else {
		queryUrl = cityurl // Constant value
	}
	params := url.Values{}
	urlToSend, err := url.Parse(queryUrl)
	if err != nil {
		return "", err
	}
	params.Add("f", "json")
	params.Add("maxSuggestions", "1")
	params.Add("text", query)
	params.Add("countryCode", "USA,PRI,VIR,GUM,ASM")
	params.Add("category", "Land Features,Bay,Channel,Cove,Dam,Delta,Gulf,Lagoon,Lake,Ocean,Reef,Reservoir,Sea,Sound,Strait,Waterfall,Wharf,Amusement Park,Historical Monument,Landmark,Tourist Attraction,Zoo,College,Beach,Campground,Golf Course,Harbor,Nature Reserve,Other Parks and Outdoors,Park,Racetrack,Scenic Overlook,Ski Resort,Sports Center,Sports Field,Wildlife Reserve,Airport,Ferry,Marina,Pier,Port,Resort,Postal,Populated Place")
	urlToSend.RawQuery = params.Encode()
	req, _ := http.NewRequest("GET", urlToSend.String(), nil)
	addHeaders(req, w.UserAgent, true)
	res, err := w.Client.Do(req)
	if err != nil {
		slog.Error("Error sending request for city lookup", "error", errors.Unwrap(err))
		return "", errors.Join(err, ErrInvalidCityResult)
	}
	defer res.Body.Close()
	response, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read results of city lookup", "error", errors.Unwrap(err))
		return "", errors.Join(err, ErrInvalidCityResult)
	}
	slog.Debug("Received response", "response", response)
	var results searchSuggestions
	if err = json.Unmarshal(response, &results); err != nil {
		slog.Error("Error decoding response of city lookup", "error", errors.Unwrap(err))
		return "", errors.Join(err, ErrInvalidCityResult)
	}
	slog.Debug("Received results from geocode.arcgis.com", "results", results)
	if len(results.Suggestions) != 1 {
		slog.Error("Received an invalid number of results when fetching city name", "zip", query, "numResults", len(results.Suggestions))
		return "", ErrInvalidCityResult
	}
	return results.Suggestions[0].Result, nil
}

// city should be the result of GetCity in the format "zip, city, state, country"
// e.g. "53226, Milwaukee, WI, USA"
func (w *getWeatherData) GetLatLong(city string) (*LatLong, error) {
	if w.Client == nil {
		w.Client = &http.Client{}
	}
	var queryUrl string
	if w.LatLongUrl != "" {
		queryUrl = w.LatLongUrl
	} else {
		queryUrl = latlongurl
	}
	params := url.Values{}
	params.Add("f", "json")
	params.Add("text", city)
	urlToSend, _ := url.Parse(queryUrl)
	urlToSend.RawQuery = params.Encode()
	req, _ := http.NewRequest("GET", urlToSend.String(), nil)
	addHeaders(req, w.UserAgent, true)
	res, err := w.Client.Do(req)
	if err != nil {
		slog.Error("Failed to get Lat/Long values", "error", errors.Unwrap(err))
		return nil, errors.Join(err, ErrInvalidLatLongResult)
	}
	defer res.Body.Close()
	response, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read results of lat/long query", "error", errors.Unwrap(err))
		return nil, errors.Join(err, ErrInvalidLatLongResult)
	}
	slog.Debug("Received response from geocode.arcgis.com", "response", response)
	var results locationResults
	if err = json.Unmarshal(response, &results); err != nil {
		slog.Error("Failed to unmarshal results of lat/log query", "error", errors.Unwrap(err))
		return nil, errors.Join(err, ErrInvalidLatLongResult)
	}
	slog.Debug("Received results from geocode.arcgis.com", "results", results)
	if len(results.Locations) != 1 {
		slog.Error("Received an invalid number of results when fetching lat/log", "city", city, "numResults", len(results.Locations))
		return nil, errors.Join(err, ErrInvalidLatLongResult)
	}
	return &LatLong{
		results.Locations[0].Extent.Ymin, // Latitude
		results.Locations[0].Extent.Xmin, // Longitude
	}, nil
}

func (w *getWeatherData) GetWeather(LatLong *LatLong) (*Weather, error) {
	if w.Client == nil {
		w.Client = &http.Client{}
	}
	var queryUrl string
	if w.WeatherUrl != "" {
		queryUrl = w.WeatherUrl
	} else {
		queryUrl = weatherurl
	}
	doc, err := downloadWeather(LatLong, queryUrl, w.UserAgent, w.Client)
	if err != nil {
		return nil, err
	}
	return parseWeather(doc)
}
