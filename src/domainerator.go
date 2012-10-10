package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
	"net/http"
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
	dnsServers  = flag.String("dns", "", "Comma-separated list of DNS servers to talk to")
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

// Load known domain prefixes and TLDs from Mozilla
func loadKnownTopLevelDomains() (tlds map[string]bool, err error) {
	tlds = make(map[string]bool)
	resp, err := http.Get(
		"http://mxr.mozilla.org/mozilla-central/source/netwerk/dns/effective_tld_names.dat?raw=1")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	lines := strings.Split(string(body), "\n")
	for _, s := range lines {
		s = strings.TrimSpace(s)
		if len(s) == 0 || strings.HasPrefix(s, "//") || strings.HasPrefix(s, "!") {
			continue
		}
		if strings.HasPrefix(s, "*.") {
			s = strings.Replace(s, "*.", "", 1)
		}
		tlds[s] = true
	}
	return
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
	domain     string
	registered bool
}

// Format DomainResult into string for output file
func (dr DomainResult) String() string {
	return fmt.Sprintf("%s\t%t\n", dr.domain, dr.registered)
}

// Returns true if domain has a Name Server associated
func hasNS(domain, dnsServer string) (bool, error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return false, err
	}
	c := new(dns.Client)
	c.Net = "tcp"
	c.ReadTimeout = time.Duration(4) * time.Second
	c.WriteTimeout = c.ReadTimeout
	c.Retry = true
	c.Attempts = 3
	m := new(dns.Msg)
	m.SetQuestion(domain+".", dns.TypeNS)
	m.RecursionDesired = true
	if len(dnsServer) == 0 {
		dnsServer = config.Servers[0] + ":" + config.Port
	}
	in, err := c.Exchange(m, dnsServer)
	if err != nil {
		return false, err
	}
	return in.Rcode == dns.RcodeSuccess, nil
}

// Check if each domain is registered
func checkDomains(in <-chan string, out chan<- DomainResult, dnsServer string) {
	for domain := range in {
		var registered bool
		var err error
		for i := 0; i < 5; i++ {
			registered, err = hasNS(domain, dnsServer)
			if err == nil {
				break
			}
			fmt.Fprintf(os.Stderr, "\nRetrying %q in DNS %q...\n", domain, dnsServer)
		}
		if err == nil {
			out <- DomainResult{domain, registered}
		} else {
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

	fmt.Print("Loading TLDs list... ")
	accepted, err := loadKnownTopLevelDomains()
	if err != nil {
		showErrorAndExit(err, 3)
	}
	fmt.Println("done.")

	fmt.Print("Loading word lists.. ")
	prefixes, err := loadWordListFile(flag.Arg(0))
	if err != nil {
		showErrorAndExit(err, 4)
	}
	suffixes, err := loadWordListFile(flag.Arg(1))
	if err != nil {
		showErrorAndExit(err, 4)
	}
	fmt.Println("done.")

	tlds, err := parseTopLevelDomains(*tldsCsv, accepted)
	if err != nil {
		showErrorAndExit(err, 5)
	}

	outputPath := flag.Arg(2)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		showErrorAndExit(err, 6)
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

	if len(*dnsServers) > 0 {
		dnses = strings.Split(*dnsServers, ",")
	}

	for i := 0; i < *concurrency; i++ {
		if len(dnses) > 0 {
			dnsServer = dnses[curDns]
			curDns += 1
			if curDns >= len(dnses) {
				curDns = 0
			}
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