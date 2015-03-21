package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/hgfischer/domainerator/domain/name"
	"github.com/hgfischer/domainerator/domain/ns"
	"github.com/hgfischer/domainerator/domain/query"
	"github.com/hgfischer/domainerator/wordlist"
)

const (
	defaultPublicSuffixes = "com,net,org,biz,info,mobi,name,tel,us,in,me,co,ca,de,eu,ws,it,be,at,ch,cz,li,nl,pl,re,so,wf,im,no"
	defaultDNSServers     = "8.8.8.8,8.8.4.4,4.2.2.1,4.2.2.2,4.2.2.3,4.2.2.4,4.2.2.5,4.2.2.6,198.153.192.1,198.153.194.1,67.138.54.100,207.225.209.66"
)

// Command line options
var (
	single      = flag.Bool("single", true, "Also check single words")
	itself      = flag.Bool("itself", false, "Include words combined with itself")
	hyphenate   = flag.Bool("hyphen", false, "Include hyphenated combinations")
	hacks       = flag.Bool("hacks", true, "Enable domain hacks")
	fuse        = flag.Bool("fuse", true, "Fuse words if letters match (ex.: ab + bc = abbc => abc")
	includeTLDs = flag.Bool("tlds", false, "Include all TLDs in public domain suffix list")
	includeUTF8 = flag.Bool("utf8", false, "Include combinations with UTF-8 characters")
	publicCSV   = flag.String("ps", defaultPublicSuffixes, "Public domain suffixes to combine with")
	dnsCSV      = flag.String("dns", defaultDNSServers, "Comma-separated list of DNS servers to talk to")
	protocol    = flag.String("proto", "udp", "Protocol (udp/tcp)")
	maxLength   = flag.Int("maxlen", 64, "Maximum length of generated domains including public suffix")
	minLength   = flag.Int("minlen", 3, "Minimum length of generated domains without public suffic")
	concurrency = flag.Int("c", 50, "Number of concurrent threads doing checks")
	available   = flag.Bool("avail", true, "If true, output only available domains (NXDOMAIN) without DNS status code")
	strictMode  = flag.Bool("strict", true, "If true, filter some possibly prohibited domains (domain == tld, etc)")
)

// Prints an error message to stderr and exist with a return code
func showErrorAndExit(err error, returnCode int) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(returnCode)
}

// Print command line help and exit application
func usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: domainerator [flags] [prefixes wordlist] [suffixes wordlist] [output file]\n")
	fmt.Fprintf(os.Stderr, "\nFlags:\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func loadWordList(file string) (list []string) {
	list, err := wordlist.Load(file)
	if err != nil {
		showErrorAndExit(err, 10)
	}
	return
}

func loadWordLists(prefixFile, suffixFile string) (prefixes, suffixes []string) {
	fmt.Print("Loading word lists.. ")
	prefixes = loadWordList(prefixFile)
	suffixes = loadWordList(suffixFile)
	if len(prefixes) == 0 && len(suffixes) == 0 {
		showErrorAndExit(errors.New("Empty wordlists"), 12)
	}
	fmt.Println("done.")
	return
}

func loadFlags() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 3 {
		fmt.Fprintf(os.Stderr, "Error: Missing some word list file path and/or output file path\n")
		flag.Usage()
	}
}

func loadPublicSuffixList() (psl []string) {
	psl, err := name.ParsePublicSuffixCSV(*publicCSV, ns.PublicSuffixes, *includeTLDs)
	if err != nil {
		showErrorAndExit(err, 20)
	}
	if !*includeUTF8 {
		psl = wordlist.FilterUTF8(psl)
	}
	fmt.Printf("Public Suffixes: %s\n", strings.Join(psl, ", "))
	return
}

func loadDNSServers() (dnsServers []string) {
	dnsServers = name.ParseDNSCSV(*dnsCSV)
	if len(dnsServers) == 0 {
		showErrorAndExit(errors.New("You need to specify a DNS server"), 30)
	}
	return
}

func checkProtocol() {
	if *protocol != "tcp" && *protocol != "udp" {
		errmsg := fmt.Sprintf("Unknown protocol: %q (should be \"udp\" or \"tcp\"", *protocol)
		showErrorAndExit(errors.New(errmsg), 35)
	}
}

func setupOutputFile(outputPath string) (outputFile *os.File) {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		showErrorAndExit(err, 40)
	}
	return
}

func createDomainList(prefixes, suffixes, psl []string) (domains []string) {
	fmt.Print("Creating domain list... ")
	domains = name.Combine(prefixes, suffixes, psl, *single, *hyphenate, *itself, *hacks, *fuse, *minLength)
	if !*includeUTF8 {
		domains = wordlist.FilterUTF8(domains)
	}
	domains = name.FilterMaxLength(domains, *maxLength)
	if *strictMode {
		domains = name.FilterStrictDomains(domains, ns.PublicSuffixes)
	}
	domains = wordlist.RemoveDuplicates(domains)
	fmt.Println("done.")
	if len(domains) == 0 {
		showErrorAndExit(errors.New("I could not generate a single valid domain"), 50)
	}
	return
}

func printFeedback(startTime time.Time, processed, total int) {
	fmtStr := "\rChecked %d of %d domains. Elapsed %s. ETA %s. Goroutines: %d\033[K"
	elapsed := time.Since(startTime)
	etaSecs := elapsed.Seconds() * float64(total) / float64(processed)
	eta := time.Duration(etaSecs) * time.Second
	out := fmt.Sprintf(fmtStr, processed, total, elapsed, eta, runtime.NumGoroutine())
	fmt.Print(out)
}

func saveDomainResult(outputFile *os.File, r query.Result, available bool) {
	if (available && r.Available()) || !available {
		_, err := outputFile.WriteString(r.String(available))
		if err != nil {
			showErrorAndExit(err, 6)
		}
	}
}

// MAIN
func main() {
	loadFlags()
	prefixes, suffixes := loadWordLists(flag.Arg(0), flag.Arg(1))
	psl := loadPublicSuffixList()
	dnsServers := loadDNSServers()
	checkProtocol()
	outputFile := setupOutputFile(flag.Arg(2))
	defer outputFile.Close()
	domains := createDomainList(prefixes, suffixes, psl)
	pending, retries, complete := make(chan string), make(chan string), make(chan query.Result)

	fmt.Println("Starting checks... ")
	startTime := time.Now()

	// start checks
	for i := 0; i < *concurrency; i++ {
		go query.CheckDomains(i, pending, retries, complete, dnsServers, *protocol)
	}

	// send domains
	go func() {
		for _, domain := range domains {
			pending <- domain
		}
		close(pending)
	}()

	// save results and print feedback
	wrote := 0
	for r := range complete {
		saveDomainResult(outputFile, r, *available)
		wrote++
		printFeedback(startTime, wrote, len(domains))
		if wrote == len(domains) {
			close(complete)
		}
	}
	fmt.Println("\nDone.")
}
