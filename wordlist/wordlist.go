// Package wordlist implements some wordlists functions
package wordlist

import (
	"io/ioutil"
	"strings"
	"unicode/utf8"
)

// TrimWords spaces for each word in a wordlist
func TrimWords(words []string) []string {
	for key, word := range words {
		words[key] = strings.Replace(word, " ", "", -1)
	}
	return words
}

// Load a word list file in a strings slice and return it
func Load(filePath string) ([]string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	words := strings.Split(string(content), "\n")
	words = TrimWords(words)
	words = FilterEmptyWords(words)
	return words, nil
}

// FilterEmptyWords remove empty words from the word list
func FilterEmptyWords(words []string) []string {
	var filtered []string
	for _, word := range words {
		if len(word) > 0 {
			filtered = append(filtered, word)
		}
	}
	return filtered
}

// RemoveDuplicates remove duplicate strings from a slice
func RemoveDuplicates(words []string) []string {
	m := map[string]bool{}
	cleaned := []string{}
	for _, str := range words {
		if _, seen := m[str]; !seen {
			cleaned = append(cleaned, str)
			m[str] = true
		}
	}
	return cleaned
}

// FilterUTF8 remove words with UTF8 encoded characters
func FilterUTF8(words []string) []string {
	var filtered []string
	for _, word := range words {
		if utf8.RuneCountInString(word) == len(word) {
			filtered = append(filtered, word)
		}
	}
	return filtered
}

// FromCSV parse a CSV string into a cleaned slice of strings
func FromCSV(csv string) []string {
	csv = strings.TrimSpace(csv)
	words := strings.Split(csv, ",")
	words = TrimWords(words)
	words = FilterEmptyWords(words)
	return words
}
