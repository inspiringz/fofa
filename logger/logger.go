package logger

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	envA         = os.Getenv("FOFA_EMAIL")
	envK         = os.Getenv("FOFA_KEY")
	pWARN        = "\033[1;31m[!]\033[0m "
	pINFO        = "\033[1;33m[+]\033[0m "
	pSUCCESS     = "\033[1;32m[*]\033[0m "
	isWin        = false
	usageCommand = "Usage:\n  ./fofa -m fofa_email -k fofa_key -q 'header=\"elastic\"' -s 10000 -o data.xlsx\n" +
		"  ./fofa -m fofa_email -k fofa_key -f query_rules_file.txt -s 10000 -o data.xlsx\n"
	usageOptions = `Options:
  -h, --help
  -m, --mail MAIL            fofa email account (default: ` + envA + `)
  -k, --key KEY              fofa api key (default: ` + envK + `)
  -q, --query QUERY          query string (default: '')
  -f, --file FILE            batch query rules file (default: '')
  -s, --size SIZE            export data volume (default: 10000)
  -o, --output OUTPUT        output filename / absolute path (default: data.xlsx)
`
)

const (
	VERSION     = 1.0
	asciiBanner = `
      ░░░░▐▐░░░  dMMMMMP .aMMMb  dMMMMMP .aMMMb 
 ▐  ░░░░░▄██▄▄  dMP     dMP"dMP dMP     dMP"dMP 
  ▀▀██████▀░░  dMMMP   dMP dMP dMMMP   dMMMMMP  
  ░░▐▐░░▐▐░░  dMP     dMP.aMP dMP     dMP dMP   
 ▒▒▒▐▐▒▒▐▐▒  dMP      VMMMP" dMP     dMP dMP
 https://github.com/inspiringz/fofa
`

	winAsciiBanner = "\n  ,-.       _,---._ __  / \\ \n" +
		" /  )    .-'       `./ /   \\ \n" +
		"(  (   ,'            `/    /|\n" +
		" \\  `-\"             \\'\\   / |\n" +
		"  `.              ,  \\ \\ /  |\n" +
		"   /`.          ,'-`----Y   |\n" +
		"  (            ;        |   '\n" +
		"  |  ,-.    ,-'         |  /\n" +
		"  |  | (   |            | /\n" +
		"  )  |  \\  `.___________|/\n" +
		"  `--'   `--'\n" +
		"https://github.com/inspiringz/fofa\n"
)

func InitPlatform() {
	if strings.HasSuffix(strings.ToLower(os.Args[0]), ".exe") {
		pWARN = "[!] "
		pINFO = "[+] "
		pSUCCESS = "[*] "
		isWin = true
	}
	usageCommand = strings.Replace(usageCommand, "./fofa", os.Args[0], 2)
	return
}

func randColor() string {
	rand.Seed(time.Now().UnixNano())
	colorNum := rand.Intn(6) + 31
	return strconv.Itoa(colorNum)
}

func AsciiBanner() {
	if isWin != true {
		fmt.Println("\033[1;" + randColor() + "m" + asciiBanner + "\033[0m")
	} else {
		fmt.Println(winAsciiBanner)
	}
}

func Usage() {
	fmt.Println(usageCommand)
	fmt.Println(usageOptions)
}

func Warn(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, pWARN+format+"\n", args...)
}

func Info(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, pINFO+format+"\n", args...)
}

func Success(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, pSUCCESS+format+"\n", args...)
}
