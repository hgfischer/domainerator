package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	PS_URL = "http://mxr.mozilla.org/mozilla-central/source/netwerk/dns/effective_tld_names.dat?raw=1"
)

// Load known public suffix list (http://publicsuffix.org)
func loadPublicSuffixList() (suffixes []string, err error) {
	resp, err := http.Get(PS_URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "!") {
			continue
		}
		if strings.HasPrefix(line, "*.") {
			line = strings.Replace(line, "*.", "", 1)
		}
		suffixes = append(suffixes, line)
	}
	return
}

func main() {
	suffixes, err := loadPublicSuffixList()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "package ns\n\n")
	fmt.Fprintf(os.Stdout, "var PublicSuffixes = map[string]bool{\n")
	for _, suffix := range suffixes {
		fmt.Fprintf(os.Stdout, "\t%q: true,\n", suffix)
	}
	fmt.Fprintf(os.Stdout, "}\n")
}
