package main

import (
	"flag"
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"errors"
)

// Command line options
var (
	verbose = flag.Bool("v", false, "Verbose mode")
	single = flag.Bool("s", false, "Also check single words")
	hyphenate = flag.Bool("H", false, "Include hyphenated combinations")
	tldsCsv = flag.String("tlds", "com,net,org", "TLDs to combine with")
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
func btou(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Calculate possible combinations 
func calculateCombinations(wordCount, tldCount int, single, hyphenate bool) (result int) {
	c := wordCount * tldCount
	return c + (c * btou(single)) + (c * btou(hyphenate))
}

// Combine words and tlds to make the ordered domain list
func combineWords(words, tlds []string, single, hyphenate bool) ([]string) {
	combinations := calculateCombinations(len(words), len(tlds), single, hyphenate)
	domains := make([]string, combinations)
	for _, prefix := range words {
		if single {
			for _, tld := range tlds {
				domains = append(domains, prefix + "." + tld)
			}
		}
		for _, suffix := range words {
			for _, tld := range tlds {
				domains = append(domains, prefix + suffix + "." + tld)
				if hyphenate {
					domains = append(domains, prefix + "-" + suffix + "." + tld)
				}
			}
		}
	}
	return domains
}

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

	domains, err := combineWords(words, tlds, *single, *hyphenate)
	fmt.Printf("%q\n", domains)

	/*
		TODO
		- solve IPs to check what domain is registered or not
			- addrs, err = net.LookupHost(domain)
		- check for whois?
		- check for MX entries?
	*/
}
