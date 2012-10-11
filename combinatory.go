package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"unicode/utf8"
)

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
func parsePublicSuffixCsv(csv string, accepted map[string]bool, includeTLDs bool) ([]string, error) {
	csv = strings.TrimSpace(csv)
	if len(csv) == 0 {
		return nil, errors.New("empty Public Suffix list")
	}
	psl := strings.Split(csv, ",")
	for k, v := range psl {
		psl[k] = strings.TrimSpace(v)
	}
	for _, ps := range psl {
		_, ok := accepted[ps]
		if !ok {
			return nil, errors.New(fmt.Sprintf("Public Suffix %q is unknown", ps))
		}
	}
	if includeTLDs {
		for ps, _ := range accepted {
			if !strings.Contains(ps, ".") {
				psl = append(psl, ps)
			}
		}
	}

	psl = removeDuplicates(psl)
	sort.Strings(psl)
	return psl, nil
}

func combineWordAndPublicSuffixes(word string, psl []string, hacks bool) []string {
	domains := make([]string, 0)
	for _, ps := range psl {
		domains = append(domains, word+"."+ps)
		if hacks {
			last := strings.LastIndex(word, ps)
			if last > 0 {
				domains = append(domains, word[:last+1]+"."+ps)
			}
		}
	}
	return domains
}

// Combine words and public suffixes to make the ordered domain list
func combineWords(prefixes, suffixes, psl []string, single, singleOnly, hyphenate, itself bool, hacks bool, maxLength int) []string {
	domains := make([]string, 0)
	if single || singleOnly {
		for _, prefix := range prefixes {
			domains = append(domains, combineWordAndPublicSuffixes(prefix, psl, hacks)...)
		}
		for _, suffix := range suffixes {
			domains = append(domains, combineWordAndPublicSuffixes(suffix, psl, hacks)...)
		}
	}
	if !singleOnly {
		for _, prefix := range prefixes {
			for _, suffix := range suffixes {
				if prefix == suffix && !itself {
					continue
				}
				domains = append(domains, combineWordAndPublicSuffixes(prefix+suffix, psl, hacks)...)
				if hyphenate {
					domains = append(domains, combineWordAndPublicSuffixes(prefix+"-"+suffix, psl, hacks)...)
				}
			}
		}
	}
	domains = removeDuplicates(domains)
	filtered := make([]string, 0)
	for _, domain := range domains {
		if len(domain) <= maxLength {
			filtered = append(filtered, domain)
		}
	}

	sort.Strings(filtered)
	return filtered
}

func filterUTF8(domains []string) []string {
	output := make([]string, 0)
	for _, domain := range domains {
		if utf8.RuneCountInString(domain) == len(domain) {
			output = append(output, domain)
		}
	}
	return output
}
