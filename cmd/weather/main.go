package main

import (
	"log/slog"
	"net/http"
	"os"

	weather "github.com/inventor500/go-weather"
)

type Args struct {
	UserAgent string
	Location  string
}

var args *Args
var printer = CreatePrinter()

func init() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	var err error
	args, err = ParseArgs()
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	os.Exit(mainFunc(printer))
}

func mainFunc(printer Printer) int {
	client := http.Client{}
	res, err := weather.GetCity(args.Location, args.UserAgent, &client)
	if err != nil {
		return 1
	}
	latlong, err := weather.GetLatLong(res, args.UserAgent, &client)
	if err != nil {
		return 1
	}
	doc, err := weather.GetWeather(latlong, args.UserAgent, &client)
	if err != nil {
		return 1
	}
	w, err := weather.ParseWeather(doc)
	if err != nil || w == nil {
		return 1
	}
	printer.Printf(Underline|Bold, "Weather Advisories\n")
	for i := 0; i < len(w.Advisories); i++ {
		printer.Printf(Red, "%s\n", w.Advisories[i].Description)
	}
	printer.Printf(Underline|Bold, "Upcoming Weather Events\n")
	for i := 0; i < len(w.WeatherTimes); i++ {
		printer.Printf(Bold, "%s ", w.WeatherTimes[i].Label)
		printer.Printf(Regular, "%s\n", w.WeatherTimes[i].LongDesc)
	}
	return 0
}
