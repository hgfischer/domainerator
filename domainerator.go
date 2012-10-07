package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

const (
	defaultConcurrency = 10 // How many goroutines should we run at the same time?
)

// Command line options
var (
	single      = flag.Bool("s", false, "Also check single words")
	onlySingle  = flag.Bool("S", false, "Check only single words combined with TLDs")
	itself      = flag.Bool("i", false, "Include words combined with itself")
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
	fmt.Fprintf(os.Stderr, "Usage: domainerator [flags] [word list file] [output file]\n")
	fmt.Fprintf(os.Stderr, "\nFlags:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

// Remove duplicate strings from a slice
func removeDuplicates(strings []string) []string {
	m := map[string]bool{}
	res := []string{}
	for _, str := range strings {
		if _, seen := m[str]; !seen {
			res = append(res, str)
			m[str] = true
		}
	}
	return res
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
	list := strings.Split(string(content), "\n")
	list = removeDuplicates(list)
	return list, nil
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
	tlds = removeDuplicates(tlds)
	return tlds, nil
}

// Combine words and tlds to make the ordered domain list
func combineWords(words, tlds []string, single, onlySingle, hyphenate, itself bool) []string {
	domains := make([]string, 0)
	for _, prefix := range words {
		if single || onlySingle {
			for _, tld := range tlds {
				domains = append(domains, prefix+"."+tld)
			}
		}
		if !onlySingle {
			for _, suffix := range words {
				if prefix == suffix && !itself {
					continue
				}
				for _, tld := range tlds {
					domains = append(domains, prefix+suffix+"."+tld)
					if hyphenate {
						domains = append(domains, prefix+"-"+suffix+"."+tld)
					}
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

// Format DomainResult into string for output file
func (dr DomainResult) String() string {
	var err string
	if dr.err == nil {
		err = ""
	} else {
		err = fmt.Sprintf("%q", dr.err)
	}
	return fmt.Sprintf("%s\t%t\t%q\t%q\n",
		dr.domain, dr.registered, dr.addresses, err)
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
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Error: Missing word list file path and/or output file path\n")
		flag.Usage()
	}

	fmt.Print("Loading word list.. ")
	words, err := loadWordListFile(flag.Arg(0))
	if err != nil {
		showErrorAndExit(err, 3)
	}
	fmt.Println("done.")

	tlds, err := parseTopLevelDomains(*tldsCsv)
	if err != nil {
		showErrorAndExit(err, 4)
	}

	outputPath := flag.Arg(1)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		showErrorAndExit(err, 5)
	}
	defer outputFile.Close()

	fmt.Print("Creating domain list... ")
	domains := combineWords(words, tlds, *single, *onlySingle, *hyphenate, *itself)
	fmt.Println("done.")

	fmt.Println("Starting checks... ")
	startTime := time.Now()
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

	digits := len(fmt.Sprintf("%d", len(domains)))
	fmtStr := fmt.Sprintf("\rChecked %%%dd of %%%dd domains. Elapsed %%10s. ETA %%8s.", digits, digits)

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
			fmt.Printf(fmtStr, processed, len(domains), elapsed, eta)
		}
	}

	fmt.Println()
}
