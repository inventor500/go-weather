package weather

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

func downloadWeather(latlong *LatLong, queryUrl, userAgent string, client *http.Client) (*goquery.Document, error) {
	if latlong == nil {
		logError("Attempted to call GetWeather with undefined latitude/longitude")
		return nil, ErrInvalidParameter
	}
	if client == nil {
		client = &http.Client{}
	}
	params := url.Values{}
	params.Add("lat", fmt.Sprint(latlong.Lat))
	params.Add("lon", fmt.Sprint(latlong.Long))
	sendUrl, _ := url.Parse(queryUrl)
	sendUrl.RawQuery = params.Encode()
	req, _ := http.NewRequest("GET", sendUrl.String(), nil)
	addHeaders(req, userAgent, false)
	res, err := client.Do(req)
	if err != nil {
		logError("Error sending request for weather: %s", errors.Unwrap(err))
		return nil, errors.Join(err, ErrWeatherParse)
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		logError("Error parsing HTML: %s", errors.Unwrap(err))
		return nil, errors.Join(err, ErrWeatherParse)
	}
	return doc, nil
}

func parseWeather(doc *goquery.Document) (*Weather, error) {
	if doc == nil {
		logError("Attempted to call ParseWeather with undefined doc")
		return nil, ErrInvalidParameter
	}
	var weather Weather
	// Without .forecast-tombstone, this only finds one element
	shortForecast := doc.Find("#seven-day-forecast-list .forecast-tombstone")
	weather.WeatherTimes = make([]WeatherTime, shortForecast.Length())
	// Get the short descriptions
	shortForecast.Each(func(i int, s *goquery.Selection) {
		weather.WeatherTimes[i] = WeatherTime{
			strings.TrimSpace(s.Find(".period-name").First().Text()), // Name
			strings.TrimSpace(s.Find(".temp").First().Text()),        // Temperature
			strings.TrimSpace(s.Find(".short-desc").First().Text()),  // Short Description
			"", // Some of these have the long description as the image alt text, but some do not
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
		weather.Advisories[i].Description = strings.TrimSpace(s.Text()) // TODO: Needs for information than this
	})
	return &weather, nil
}
