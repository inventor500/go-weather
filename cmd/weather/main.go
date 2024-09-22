package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	weather "github.com/inventor500/go-weather"
)

type Args struct {
	Zip string
}

var args Args

func init() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	if len(os.Args) != 2 { // TODO: More involved arguments, including °C/F, the user agent, etc
		fmt.Fprintf(os.Stderr, "Useage: %s <zip>\n", os.Args[0])
		os.Exit(1)
	}
	args.Zip = os.Args[0]
}

func main() {
	os.Exit(mainFunc())
}

func mainFunc() int {
	// TODO: Don't hard-code this
	const userAgent = "Mozilla/5.0 (Windows NT 10.0; rv:130.0) Gecko/20100101 Firefox/130.0"
	client := http.Client{}
	res, err := weather.GetCity(args.Zip, userAgent, &client)
	if err != nil {
		return 1
	}
	latlong, err := weather.GetLatLong(res, userAgent, &client)
	if err != nil {
		return 1
	}
	doc, err := weather.GetWeather(latlong, userAgent, &client)
	if err != nil {
		return 1
	}
	w, err := weather.ParseWeather(doc)
	if err != nil || w == nil {
		return 1
	}
	// TODO: Format better, use bold when this is going to a terminal
	fmt.Println("Upcoming Weather Events\n-----------------------")
	for i := 0; i < len(w.WeatherTimes); i++ {
		fmt.Println(w.WeatherTimes[i])
	}
	return 0
}
