package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// The Freedesktop standard place for a config file
const defaultConfigFile = "$XDG_CONFIG_HOME/weather/config.json"

// Firefox 130 on Windows 10
const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; rv:130.0) Gecko/20100101 Firefox/130.0"

func readConfig(config string) (*Config, error) {
	var configFile Config
	if config != "" {
		file, err := os.Open(os.ExpandEnv(config))
		if err != nil {
			if config != defaultConfigFile {
				fmt.Fprintf(os.Stderr, "Unable to read config file %s\n", config)
				return nil, errors.New("unable to read file")
			} else {
				return nil, nil
			}
		}
		defer file.Close()
		if err = json.NewDecoder(file).Decode(&configFile); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse config file %s: %s\n", config, err)
			return nil, errors.New("unable to parse file")
		}
	}
	return &configFile, nil
}

func ParseArgs() (*Args, error) {
	userAgent := flag.String("u", "", "The user agent to use for requests")
	location := flag.String("zip", "", "The location to query for")
	config := flag.String("c", defaultConfigFile, "Config file location")
	flag.Parse()
	configFile, err := readConfig(*config)
	if err != nil {
		return nil, err
	}
	// User agents are from most specific to least specific
	// i.e. specific to this run, to this program, to this user/system
	// and then use a default if nothing else works
	if *userAgent == "" {
		if configFile != nil && configFile.UserAgent != "" { // Check the config file
			*userAgent = configFile.UserAgent
		} else if os.Getenv("USER_AGENT") != "" { // Check the environment
			*userAgent = os.Getenv("USER_AGENT")
		} else { // Use a default
			*userAgent = defaultUserAgent
		}
	}
	if *location == "" {
		if len(flag.Args()) == 1 { // Check if there are any unprocessed arguments
			*location = flag.Args()[0]
		} else if configFile != nil && configFile.Location != "" {
			*location = configFile.Location
		} else { // No location defined
			fmt.Fprintf(os.Stderr, "No location provided!\nUsage:\n")
			flag.PrintDefaults()
			return nil, errors.New("no arguments")
		}
	}
	return &Args{
		*userAgent,
		*location,
	}, nil
}

const Regular = 0
const Bold = 0b1
const Italic = 0b10
const Underline = 0b100
const Red = 0b1000
const Yellow = 0b10000

type Printer interface {
	Printf(int, string, ...any) (int, error)
	Fprintf(int, io.Writer, string, ...any) (int, error)
}

type print struct {
	isTerminal bool
}

func (p *print) Printf(format int, s string, a ...any) (int, error) {
	s = makeFormat(format, s, p.isTerminal)
	return fmt.Printf(s, a...)
}

func (p *print) Fprintf(format int, w io.Writer, s string, a ...any) (int, error) {
	s = makeFormat(format, s, p.isTerminal)
	return fmt.Fprintf(w, s, a...)
}

func CreatePrinter() Printer {
	stat, _ := os.Stdout.Stat()
	return &print{
		((stat.Mode() & os.ModeCharDevice) == os.ModeCharDevice),
	}
}

// TODO: This does not seem like a good solution
func makeFormat(f int, s string, i bool) string {
	if f == Regular || !i { // Nothing to do here
		return s
	}
	builder := strings.Builder{}
	builder.WriteByte('\x1B')
	builder.WriteByte('[')
	addByte(&f, &builder, Bold, []byte{'1'})
	addByte(&f, &builder, Italic, []byte{'3'})
	addByte(&f, &builder, Underline, []byte{'4'})
	addByte(&f, &builder, Red, []byte{'3', '1'})
	addByte(&f, &builder, Yellow, []byte{'3', '7'})
	builder.WriteByte('m')
	builder.WriteString(s)
	builder.WriteString("\x1B[0;0m")
	return builder.String()
}

func addByte(f *int, s *strings.Builder, test int, val []byte) {
	if (*f & test) == test {
		s.Write(val)
		*f ^= test
		if *f != 0 {
			s.WriteByte(';')
		}
	}
}
