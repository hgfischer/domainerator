package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	defaultPublicSuffixes = "com,co,net,us,org,info,tv,in,im,me,ca,pl,mobi,eu,so,be,de,ws,li,at,nl,re,ca,es,it,nl,cc,tk,io"
	defaultDNSServers     = "8.8.8.8,8.8.4.4,4.2.2.1,4.2.2.2,4.2.2.3,4.2.2.4,4.2.2.5,4.2.2.6,198.153.192.1,198.153.194.1,67.138.54.100,207.225.209.66"
	defaultConcurrency    = 10
)

// Command line options
var (
	single      = flag.Bool("s", false, "Also check single words")
	singleOnly  = flag.Bool("S", false, "Check ONLY single words")
	itself      = flag.Bool("i", false, "Include words combined with itself")
	hyphenate   = flag.Bool("H", false, "Include hyphenated combinations")
	hacks       = flag.Bool("k", false, "Enable domain hacks")
	includeTLDs = flag.Bool("tlds", false, "Include all TLDs in public domain suffix list")
	skipUTF8    = flag.Bool("no-utf8", true, "Skip combinations with UTF-8 characters")
	publicCsv   = flag.String("psl", defaultPublicSuffixes, "Public domain suffixes to combine with")
	dnsServers  = flag.String("dns", defaultDNSServers, "Comma-separated list of DNS servers to talk to")
	maxLength   = flag.Int("L", 64, "Maximum length of generated domains")
	concurrency = flag.Int("c", defaultConcurrency, "Number of concurrent threads doing checks")
)

// Prints an error message to stderr and exist with a return code
func showErrorAndExit(err error, returnCode int) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(returnCode)
}

// Print command line help and exit application
func usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: domainerator [flags] [prefixes word list] [suffixes word list] [output file]\n")
	fmt.Fprintf(os.Stderr, "\nFlags:\n")
	flag.PrintDefaults()
	os.Exit(1)
}

// MAIN
func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 3 {
		fmt.Fprintf(os.Stderr, "Error: Missing some word list file path and/or output file path\n")
		flag.Usage()
	}

	fmt.Print("Loading word lists.. ")
	prefixes, err := loadWordListFile(flag.Arg(0))
	if err != nil {
		showErrorAndExit(err, 10)
	}
	suffixes, err := loadWordListFile(flag.Arg(1))
	if err != nil {
		showErrorAndExit(err, 11)
	}
	fmt.Println("done.")

	psl, err := parsePublicSuffixCsv(*publicCsv, publicSuffixes, *includeTLDs)
	if err != nil {
		showErrorAndExit(err, 20)
	}
	if *skipUTF8 {
		psl = filterUTF8(psl)
	}

	fmt.Printf("Public Suffixes: %s\n", strings.Join(psl, ", "))

	if len(strings.TrimSpace(*dnsServers)) == 0 {
		showErrorAndExit(errors.New("You need to specify a DNS server"), 30)
	}

	outputPath := flag.Arg(2)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		showErrorAndExit(err, 40)
	}
	defer outputFile.Close()

	fmt.Print("Creating domain list... ")
	domains := combineWords(prefixes, suffixes, psl, *single, *singleOnly, *hyphenate, *itself, *hacks, *maxLength)
	if *skipUTF8 {
		domains = filterUTF8(domains)
	}
	fmt.Println("done.")

	fmt.Println("Starting checks... ")
	startTime := time.Now()
	pending, complete := make(chan string), make(chan DomainResult)
	var dnses []string
	dnsServer := ""
	curDns := 0

	dnses = strings.Split(*dnsServers, ",")
	for i := 0; i < *concurrency; i++ {
		dnsServer = dnses[curDns]
		curDns += 1
		if curDns >= len(dnses) {
			curDns = 0
		}
		go checkDomains(pending, complete, dnsServer)
	}

	go func() {
		for _, domain := range domains {
			pending <- domain
		}
		close(pending)
	}()

	fmtStr := "\rChecked %d of %d domains. Elapsed %s. ETA %s. Goroutines: %d"
	outLen := 0
	processed := 0
	for r := range complete {
		processed += 1
		_, err := outputFile.WriteString(r.String())
		if err != nil {
			showErrorAndExit(err, 6)
		}
		if processed == len(domains) {
			break
		}
		if processed%10 == 0 {
			elapsed := time.Since(startTime)
			etaSecs := elapsed.Seconds() * float64(len(domains)) / float64(processed)
			eta := time.Duration(etaSecs) * time.Second
			out := fmt.Sprintf(fmtStr, processed, len(domains), elapsed, eta, runtime.NumGoroutine())
			fmt.Print(out)
			if len(out) < outLen {
				fmt.Print(strings.Repeat(" ", outLen-len(out)))
			} else {
				outLen = len(out)
			}
		}
	}

	fmt.Println("\nDone.")
}
