package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"
)

// Load known domain prefixes and TLDs from Mozilla
func loadKnownTopLevelDomains() (tlds []string, largest int, err error) {
	resp, err := http.Get("http://mxr.mozilla.org/mozilla-central/source/netwerk/dns/effective_tld_names.dat?raw=1")
	if err != nil {
		return nil, 0, err
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
		length := utf8.RuneCountInString(line)
		if length > largest {
			largest = length
		}
		tlds = append(tlds, line)
	}
	return
}

func main() {
	tlds, largest, err := loadKnownTopLevelDomains()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "package tld\n\n")
	fmt.Fprintf(os.Stdout, "var List = map[string]bool{\n")
	for _, tld := range tlds {
		fmt.Fprintf(os.Stdout, "\t%q:%strue,\n", tld, strings.Repeat(" ", largest-utf8.RuneCountInString(tld)))
	}
	fmt.Fprintf(os.Stdout, "}\n")
}
