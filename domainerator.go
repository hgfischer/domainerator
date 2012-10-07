package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

const (
	defaultConcurrency = 10 // How many goroutines should we run at the same time?
)

// Command line options
var (
	verbose     = flag.Bool("v", false, "Verbose mode")
	single      = flag.Bool("s", false, "Also check single words")
	hyphenate   = flag.Bool("H", false, "Include hyphenated combinations")
	tldsCsv     = flag.String("tlds", "com,net,org", "TLDs to combine with")
	concurrency = flag.Int("c", defaultConcurrency, "Number of concurrent threads doing checks")
)

// Prints an error message to stderr and exist with a return code
func showErrorAndExit(err error, returnCode int) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(returnCode)
}

// Print command line help and exit application
func usage() {
	fmt.Fprintf(os.Stderr, "Usage: domainerator [flags] [word list file]\n")
	fmt.Fprintf(os.Stderr, "\nFlags:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

// Load a word list file in a strings slice and return it
func loadWordListFile(filePath string) ([]string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	strContent := strings.TrimSpace(string(content))
	if len(strContent) == 0 {
		return nil, errors.New("empty word list")
	}
	return strings.Split(string(content), "\n"), nil
}

// Parse a CSV string into a cleaned slice of strings
func parseTopLevelDomains(tldCsv string) ([]string, error) {
	tldCsv = strings.TrimSpace(tldCsv)
	if len(tldCsv) == 0 {
		return nil, errors.New("empty TLD list")
	}
	tlds := strings.Split(tldCsv, ",")
	for k, v := range tlds {
		tlds[k] = strings.TrimSpace(v)
	}
	return tlds, nil
}

// Boolean to Integer
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Calculate possible combinations 
func calculateCombinations(wordCount, tldCount int, single, hyphenate bool) (result int) {
	common := (wordCount * wordCount * tldCount)
	singles := (btoi(single) * wordCount * tldCount)
	hyphenated := (btoi(hyphenate) * wordCount * wordCount * tldCount)
	return common + singles + hyphenated
}

// Combine words and tlds to make the ordered domain list
func combineWords(words, tlds []string, single, hyphenate bool) []string {
	combinations := calculateCombinations(len(words), len(tlds), single, hyphenate)
	domains := make([]string, 0, combinations)
	for _, prefix := range words {
		if single {
			for _, tld := range tlds {
				domains = append(domains, prefix+"."+tld)
			}
		}
		for _, suffix := range words {
			for _, tld := range tlds {
				domains = append(domains, prefix+suffix+"."+tld)
				if hyphenate {
					domains = append(domains, prefix+"-"+suffix+"."+tld)
				}
			}
		}
	}

	return domains
}

type DomainResult struct {
	domain     string
	registered bool
	addresses  []string
	err        error
}

// Check if each domain is registered
func checkDomains(in <-chan string, out chan<- DomainResult) {
	for domain := range in {
		addr, err := net.LookupHost(domain)
		registered := err == nil
		out <- DomainResult{domain, registered, addr, err}
	}
}

// MAIN
func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Error: Missing word list file path\n")
		flag.Usage()
	}

	words, err := loadWordListFile(flag.Arg(0))
	if err != nil {
		showErrorAndExit(err, 3)
	}

	tlds, err := parseTopLevelDomains(*tldsCsv)
	if err != nil {
		showErrorAndExit(err, 4)
	}

	domains := combineWords(words, tlds, *single, *hyphenate)

	pending, complete := make(chan string), make(chan DomainResult)

	for i := 0; i < *concurrency; i++ {
		go checkDomains(pending, complete)
	}

	go func() {
		for _, domain := range domains {
			pending <- domain
		}
		close(pending)
	}()

	processed := 0
	for r := range complete {
		processed += 1
		if !r.registered {
			fmt.Println(r.domain)
		}
		if processed == len(domains) {
			break
		}
	}
}
