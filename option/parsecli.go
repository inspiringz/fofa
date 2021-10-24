package option

import (
	"errors"
	"os"
)

var (
	PrintUsage        = errors.New("")
	PrintGrammar      = errors.New("Grammar")
	IconHash          = errors.New("IconHash")
	FofaTip           = errors.New("FofaTip")
	errorNoQueryFound = errors.New("No query statement or rule file found, -q or -f parameter must be specified.")
)

const (
	DefaultEmail  = ""
	DefaultAPIKey = ""
	DefaultSize   = "10000"
	DefaultOutput = "data.xlsx"
)

func ParseCli(args []string) (
	email string,
	apiKey string,
	query string,
	ruleFile string,
	size string,
	output string,
	err error) {

	if len(args) == 0 {
		err = PrintUsage
		return
	}

	for pos := 0; pos < len(args); pos++ {
		switch args[pos] {
		case "-m", "--mail":
			if pos+1 < len(args) {
				email = args[pos+1]
				pos++
			}
		case "-k", "--key":
			if pos+1 < len(args) {
				apiKey = args[pos+1]
				pos++
			}
		case "-q", "--query":
			if pos+1 < len(args) {
				query = args[pos+1]
				pos++
			}
		case "-f", "--file":
			if pos+1 < len(args) {
				ruleFile = args[pos+1]
				pos++
			}
		case "-s", "--size":
			if pos+1 < len(args) {
				size = args[pos+1]
				pos++
			}
		case "-o", "--output":
			if pos+1 < len(args) {
				output = args[pos+1]
				pos++
			}
		case "-t", "--tip":
			if pos+1 < len(args) {
				FofaTip = errors.New(args[pos+1])
				err = FofaTip
				return
			}
		case "-ih", "--iconhash":
			if pos+1 < len(args) {
				IconHash = errors.New(args[pos+1])
				err = IconHash
				return
			}
		case "-h", "--help":
			err = PrintUsage
			return
		case "-g", "--grammar":
			err = PrintGrammar
			return
		}
	}

	if query == "" && ruleFile == "" {
		err = errorNoQueryFound
		return
	}

	if email == "" {
		if os.Getenv("FOFA_EMAIL") != "" && os.Getenv("FOFA_KEY") != "" {
			email = os.Getenv("FOFA_EMAIL")
			apiKey = os.Getenv("FOFA_KEY")
		} else {
			email = DefaultEmail
			apiKey = DefaultAPIKey
		}
	}

	if size == "" {
		size = DefaultSize
	}

	if output == "" {
		output = DefaultOutput
	}

	return
}
