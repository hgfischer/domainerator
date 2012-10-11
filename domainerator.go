package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
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
	tldsCsv     = flag.String("tlds", "com,net,org,ca,co,cc,in,us,me,io,ws,de,eu", "TLDs to combine with")
	dnsServers  = flag.String("dns", "8.8.8.8,8.8.4.4,4.2.2.1,4.2.2.2,4.2.2.3,4.2.2.4,4.2.2.5,4.2.2.6", "Comma-separated list of DNS servers to talk to")
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
	words := make([]string, 0)
	for _, w := range list {
		w = strings.TrimSpace(w)
		if len(w) > 0 {
			words = append(words, w)
		}
	}
	return words, nil
}

// Parse a CSV string into a cleaned slice of strings
func parseTopLevelDomains(tldCsv string, accepted map[string]bool) ([]string, error) {
	tldCsv = strings.TrimSpace(tldCsv)
	if len(tldCsv) == 0 {
		return nil, errors.New("empty TLD list")
	}
	tlds := strings.Split(tldCsv, ",")
	for k, v := range tlds {
		tlds[k] = strings.TrimSpace(v)
	}
	tlds = removeDuplicates(tlds)
	for _, tld := range tlds {
		_, ok := accepted[tld]
		if !ok {
			return nil, errors.New(fmt.Sprintf("TLD %q is unknown", tld))
		}
	}

	return tlds, nil
}

// Combine words and tlds to make the ordered domain list
func combineWords(prefixes, suffixes, tlds []string, single, onlySingle, hyphenate, itself bool) []string {
	domains := make([]string, 0)
	if single || onlySingle {
		for _, prefix := range prefixes {
			for _, tld := range tlds {
				domains = append(domains, prefix+"."+tld)
			}
		}
		for _, suffix := range suffixes {
			for _, tld := range tlds {
				domains = append(domains, suffix+"."+tld)
			}
		}
	}

	if !onlySingle {
		for _, prefix := range prefixes {
			for _, suffix := range suffixes {
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

	domains = removeDuplicates(domains)
	return domains
}

type DomainResult struct {
	domain string
	rCode  int
}

// Format DomainResult into string for output file
func (dr DomainResult) String() string {
	return fmt.Sprintf("%s\t%s\n", dr.domain, dns.Rcode_str[dr.rCode])
}

// Returns true if domain has a Name Server associated
func queryNS(domain, dnsServer string) (int, error) {
	c := new(dns.Client)
	c.Net = "tcp"
	c.ReadTimeout = time.Duration(4) * time.Second
	c.WriteTimeout = c.ReadTimeout
	c.Retry = true
	c.Attempts = 3
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	m.RecursionDesired = true
	in, err := c.Exchange(m, dnsServer+":53")
	if err != nil {
		return dns.RcodeNameError, err
	}
	return in.Rcode, err
}

// Check if each domain 
func checkDomains(in <-chan string, out chan<- DomainResult, dnsServer string) {
	for domain := range in {
		var rCode int
		var err error
		// try 5 times before failing
		for i := 0; i < 5; i++ {
			rCode, err = queryNS(domain, dnsServer)
			if err == nil {
				break
			}
		}
		if err == nil {
			out <- DomainResult{domain, rCode}
		} else {
			fmt.Fprintf(os.Stderr, "\nFailed to check domain %q at DNS %q (%q)\n!", domain, dnsServer, err)
			break
		}
	}
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

	tlds, err := parseTopLevelDomains(*tldsCsv, acceptedTLDs)
	if err != nil {
		showErrorAndExit(err, 20)
	}

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
	domains := combineWords(prefixes, suffixes, tlds, *single, *onlySingle, *hyphenate, *itself)
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
