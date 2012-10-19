package name

import (
	"errors"
	"fmt"
	"github.com/hgfischer/golib/wordlist"
	"sort"
	"strings"
	"unicode/utf8"
)

// Parse a CSV string into a cleaned slice of strings
func ParsePublicSuffixCSV(csv string, accepted map[string]bool, includeTLDs bool) ([]string, error) {
	psl := wordlist.FromCSV(csv)
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
	psl = wordlist.RemoveDuplicates(psl)
	sort.Strings(psl)
	return psl, nil
}

// Parse a CSV string into a cleaned slice of strings
func ParseDNSCSV(csv string) []string {
	psl := wordlist.FromCSV(csv)
	psl = wordlist.RemoveDuplicates(psl)
	sort.Strings(psl)
	return psl
}

// Combine phrases (combined words) with public suffixes, with out without domain hacks and return a slice of strings
func CombinePhraseAndPublicSuffixes(word string, psl []string, hacks bool) []string {
	domains := make([]string, 0)
	for _, ps := range psl {
		domains = append(domains, word+"."+ps)
		if hacks {
			if strings.HasSuffix(word, ps) {
				last := strings.LastIndex(word, ps)
				if last > 0 {
					domains = append(domains, word[:last]+"."+ps)
				}
			}
		}
	}
	return domains
}

// Combine two words in all possible combinations with out without hyphenation. A suffix never comes before the prefix.
func CombinePrefixAndSuffix(prefix, suffix string, itself, hyphenate bool, minLength int) []string {
	output := make([]string, 0)
	if prefix == suffix && !itself {
		return output
	}
	str := prefix + suffix
	if len(str) >= minLength {
		output = append(output, prefix+suffix)
	}
	if hyphenate {
		str := prefix + "-" + suffix
		if len(str) >= minLength {
			output = append(output, prefix+"-"+suffix)
		}
	}
	return output
}

// Combine words and public suffixes to make the ordered domain list
func Combine(prefixes, suffixes, psl []string, single, hyphenate, itself, hacks bool, minLength int) []string {
	domains := make([]string, 0)
	if single {
		for _, prefix := range prefixes {
			domains = append(domains, CombinePhraseAndPublicSuffixes(prefix, psl, hacks)...)
		}
		for _, suffix := range suffixes {
			domains = append(domains, CombinePhraseAndPublicSuffixes(suffix, psl, hacks)...)
		}
	}
	for _, prefix := range prefixes {
		for _, suffix := range suffixes {
			phrases := CombinePrefixAndSuffix(prefix, suffix, itself, hyphenate, minLength)
			for _, phrase := range phrases {
				domains = append(domains, CombinePhraseAndPublicSuffixes(phrase, psl, hacks)...)
			}
		}
	}
	return domains
}

// Filter domains surpasing the maxLengh limit. 
func FilterMaxLength(domains []string, maxLength int) []string {
	output := make([]string, 0)
	for _, domain := range domains {
		if utf8.RuneCountInString(domain) <= maxLength {
			output = append(output, domain)
		}
	}
	return output
}

// Filter out domains possibly forbidden by registrars
func FilterStrictDomains(domains []string, publicSuffixes map[string]bool) []string {
	output := make([]string, 0)
	for _, domain := range domains {
		first := strings.Index(domain, ".")
		cleanedDomain := domain[:first]
		if _, ok := publicSuffixes[cleanedDomain]; !ok {
			output = append(output, domain)
		}
	}
	return output
}
