package main

import (
	"log/slog"
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
	getter := weather.MakeWeatherGetter(args.UserAgent)
	city, err := getter.GetCity(args.Location)
	if err != nil {
		return 1
	}
	latlong, err := getter.GetLatLong(city)
	if err != nil {
		return 1
	}
	w, err := getter.GetWeather(latlong)
	if err != nil {
		return 1
	}
	if w.Advisories.Length() > 0 {
		printer.Printf(Underline|Bold|Yellow, "Weather Advisories\n")
		printer.Printf(Red, "%s\n", w.Advisories)
		printer.Printf(Underline|Yellow, w.Advisories.Url+"\n\n")
	}
	printer.Printf(Underline|Bold, "Upcoming Weather Events\n")
	for i := 0; i < len(w.WeatherTimes); i++ {
		printer.Printf(Bold, "%s ", w.WeatherTimes[i].Label)
		printer.Printf(Regular, "%s\n", w.WeatherTimes[i].LongDesc)
	}
	return 0
}
