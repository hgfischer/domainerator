package name

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/hgfischer/domainerator/wordlist"
)

// ParsePublicSuffixCSV parse a CSV string into a cleaned slice of strings
func ParsePublicSuffixCSV(csv string, accepted map[string]bool, includeTLDs bool) ([]string, error) {
	psl := wordlist.FromCSV(csv)
	for _, ps := range psl {
		_, ok := accepted[ps]
		if !ok {
			return nil, fmt.Errorf("Public Suffix %q is unknown", ps)
		}
	}
	if includeTLDs {
		for ps := range accepted {
			if !strings.Contains(ps, ".") {
				psl = append(psl, ps)
			}
		}
	}
	psl = wordlist.RemoveDuplicates(psl)
	sort.Strings(psl)
	return psl, nil
}

// ParseDNSCSV parse a CSV string into a cleaned slice of strings
func ParseDNSCSV(csv string) []string {
	psl := wordlist.FromCSV(csv)
	psl = wordlist.RemoveDuplicates(psl)
	sort.Strings(psl)
	return psl
}

// CombinePhraseAndPublicSuffixes combine phrases (combined words) with public suffixes, with out without domain hacks
// and return a slice of strings
func CombinePhraseAndPublicSuffixes(word string, psl []string, hacks bool) []string {
	var domains []string
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

// CombinePrefixAndSuffix combine two words in all possible combinations with out without hyphenation. A suffix never
// comes before the prefix.
func CombinePrefixAndSuffix(prefix, suffix string, itself, hyphenate, fuse bool, minLength int) []string {
	var output []string
	if prefix == suffix && !itself {
		return output
	}
	str := prefix + suffix
	if len(str) >= minLength {
		output = append(output, str)
	}
	if fuse {
		if strings.HasSuffix(prefix, suffix[0:1]) {
			output = append(output, prefix+suffix[1:])
		}
		if strings.HasSuffix(prefix, suffix[0:2]) {
			output = append(output, prefix+suffix[2:])
		}
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
func Combine(prefixes, suffixes, psl []string, single, hyphenate, itself, hacks, fuse bool, minLength int) []string {
	var domains []string
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
			phrases := CombinePrefixAndSuffix(prefix, suffix, itself, hyphenate, fuse, minLength)
			for _, phrase := range phrases {
				domains = append(domains, CombinePhraseAndPublicSuffixes(phrase, psl, hacks)...)
			}
		}
	}
	return domains
}

// FilterMaxLength filter domains surpasing the maxLengh limit.
func FilterMaxLength(domains []string, maxLength int) []string {
	var output []string
	for _, domain := range domains {
		if utf8.RuneCountInString(domain) <= maxLength {
			output = append(output, domain)
		}
	}
	return output
}

// FilterStrictDomains filter out domains possibly forbidden by registrars
func FilterStrictDomains(domains []string, publicSuffixes map[string]bool) []string {
	var output []string
	for _, domain := range domains {
		first := strings.Index(domain, ".")
		cleanedDomain := domain[:first]
		if _, ok := publicSuffixes[cleanedDomain]; !ok {
			output = append(output, domain)
		}
	}
	return output
}
