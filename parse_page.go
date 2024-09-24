package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// Used to get the city name from the zip
const cityurl = "https://geocode.arcgis.com/arcgis/rest/services/World/GeocodeServer/suggest"
const latlongurl = "https://geocode.arcgis.com/arcgis/rest/services/World/GeocodeServer/find"
const weatherurl = "https://forecast.weather.gov/MapClick.php"

var ErrInvalidCityResult = errors.New("invalid city result")
var ErrInvalidLatLongResult = errors.New("invalid lat/long result")
var ErrWeatherParse = errors.New("unable to parse weather results")
var ErrInvalidParameter = errors.New("invalid parameters")

func addHeaders(req *http.Request, userAgent string, isScript bool) {
	req.Header.Set("DNT", "1")                  // Do not track
	req.Header.Set("Sec-GPC", "1")              // Global Privacy Control
	req.Header.Set("User-Agent", userAgent)     // User agent
	req.Header.Set("Pragma", "no-cache")        // Don't use the cache
	req.Header.Set("Cache-Control", "no-cache") // Don't use the cache
	if !isScript {                              // This is supposed to the user doing something
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "same-site")
		req.Header.Set("Sec-Fetch-User", "?1")
		req.Header.Set("Priority", "u=0, i") // See RFC 9218, this header is set by the web browser if browsing the site normally
	}
}

func GetCity(location string, userAgent string, client *http.Client) (string, error) {
	if client == nil {
		client = &http.Client{}
	}
	params := url.Values{}
	urlToSend, _ := url.Parse(cityurl)
	params.Add("f", "json")
	params.Add("maxSuggestions", "1")

	params.Add("text", location)
	params.Add("countryCode", "USA,PRI,VIR,GUM,ASM")
	params.Add("category", "Land Features,Bay,Channel,Cove,Dam,Delta,Gulf,Lagoon,Lake,Ocean,Reef,Reservoir,Sea,Sound,Strait,Waterfall,Wharf,Amusement Park,Historical Monument,Landmark,Tourist Attraction,Zoo,College,Beach,Campground,Golf Course,Harbor,Nature Reserve,Other Parks and Outdoors,Park,Racetrack,Scenic Overlook,Ski Resort,Sports Center,Sports Field,Wildlife Reserve,Airport,Ferry,Marina,Pier,Port,Resort,Postal,Populated Place")
	urlToSend.RawQuery = params.Encode()
	req, _ := http.NewRequest("GET", urlToSend.String(), nil)
	addHeaders(req, userAgent, true)
	res, err := client.Do(req)
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
	var results SearchSuggestions
	if err = json.Unmarshal(response, &results); err != nil {
		slog.Error("Error decoding response of city lookup", "error", errors.Unwrap(err))
		return "", errors.Join(err, ErrInvalidCityResult)
	}
	slog.Debug("Received results from geocode.arcgis.com", "results", results)
	if len(results.Suggestions) != 1 {
		slog.Error("Received an invalid number of results when fetching city name", "zip", location, "numResults", len(results.Suggestions))
		return "", ErrInvalidCityResult
	}
	return results.Suggestions[0].Result, nil
}

// city should be the result of GetCity in the format "zip, city, state, country"
// e.g. "53226, Milwaukee, WI, USA"
func GetLatLong(city, userAgent string, client *http.Client) (*LatLong, error) {
	if client == nil {
		client = &http.Client{}
	}
	params := url.Values{}
	params.Add("f", "json")
	params.Add("text", city)
	urlToSend, _ := url.Parse(latlongurl)
	urlToSend.RawQuery = params.Encode()
	req, _ := http.NewRequest("GET", urlToSend.String(), nil)
	addHeaders(req, userAgent, true)
	res, err := client.Do(req)
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
	var results LocationResults
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

func GetWeather(latlong *LatLong, userAgent string, client *http.Client) (*goquery.Document, error) {
	if latlong == nil {
		slog.Error("Attempted to call GetWeather with undefined latitude/longitude")
		return nil, ErrInvalidParameter
	}
	if client == nil {
		client = &http.Client{}
	}
	params := url.Values{}
	params.Add("lat", fmt.Sprint(latlong.Lat))
	params.Add("lon", fmt.Sprint(latlong.Long))
	sendUrl, _ := url.Parse(weatherurl)
	sendUrl.RawQuery = params.Encode()
	req, _ := http.NewRequest("GET", sendUrl.String(), nil)
	addHeaders(req, userAgent, false)
	res, err := client.Do(req)
	if err != nil {
		slog.Error("Error sending request for weather", "error", errors.Unwrap(err))
		return nil, errors.Join(err, ErrWeatherParse)
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		slog.Error("Error parsing HTML", "error", errors.Unwrap(err))
		return nil, errors.Join(err, ErrWeatherParse)
	}
	return doc, nil
}

func ParseWeather(doc *goquery.Document) (*Weather, error) {
	if doc == nil {
		slog.Error("Attempted to call ParseWeather with undefined doc")
		return nil, ErrInvalidParameter
	}
	var weather Weather
	// Without .forecast-tombstone, this only finds one element
	shortForecast := doc.Find("#seven-day-forecast-list .forecast-tombstone")
	weather.WeatherTimes = make([]WeatherTime, shortForecast.Length())
	// Get the short descriptions
	shortForecast.Each(func(i int, s *goquery.Selection) {
		weather.WeatherTimes[i] = WeatherTime{
			s.Find(".period-name").First().Text(), // Name
			s.Find(".temp").First().Text(),        // Temperature
			s.Find(".short-desc").First().Text(),  // Short Description
			"",                                    // Some of these have the long description as the image alt text, but some do not
		}
	})
	doc.Find("#detailed-forecast-body .row-forecast .forecast-text").Each(func(i int, s *goquery.Selection) {
		if i < len(weather.WeatherTimes) {
			weather.WeatherTimes[i].LongDesc = s.Text()
		}
	})
	advisories := doc.Find(".panel-danger .panel-body ul li")
	weather.Advisories = make([]Advisory, advisories.Length())
	advisories.Each(func(i int, s *goquery.Selection) {
		weather.Advisories[i].Description = s.Text() // TODO: Needs for information than this
	})
	return &weather, nil
}
