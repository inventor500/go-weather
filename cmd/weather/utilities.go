package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func ParseArgs() *Args {
	userAgent := flag.String("u", "", "The user agent to use for requests")
	zip := flag.String("zip", "", "The zip code to query for")
	flag.Parse()
	if *userAgent == "" {
		*userAgent = os.Getenv("USER_AGENT")
	}
	if *userAgent == "" { // No user agent was specified
		// Default for Firefox 130
		*userAgent = "Mozilla/5.0 (Windows NT 10.0; rv:130.0) Gecko/20100101 Firefox/130.0"
	}
	if *zip == "" {
		if len(flag.Args()) == 1 { // Use user agent from the environment
			*zip = flag.Args()[0]
		} else { // No user agent defined
			fmt.Fprintf(os.Stderr, "Usage:\n")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}
	return &Args{
		*userAgent,
		*zip,
	}
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
	if f == 0 || !i { // Nothing to do here
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
