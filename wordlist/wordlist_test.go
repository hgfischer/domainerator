package wordlist

import (
	"github.com/hgfischer/golib/tests"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestTrimWords(t *testing.T) {
	words := []string{" word ", "word  ", "    word", "wor d", "   wo rd ", "   wo    rd "}
	expected := []string{"word", "word", "word", "word", "word", "word"}
	words = TrimWords(words)
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ERR_FMT_EXPECTED_GOT, "TrimWords", expected, words)
	}
}

func setupTestLoad(words []string) (string, error) {
	file, err := ioutil.TempFile("/tmp", "domainerator.wordlist.test.")
	if err != nil {
		return "", err
	}
	defer file.Close()
	for _, word := range words {
		file.WriteString(word + "\n")
	}
	return file.Name(), nil
}

func TestLoad(t *testing.T) {
	words := []string{" word ", "word  ", "    word", "wor d", "   wo rd ", "   wo    rd ", "     ", ""}
	expected := []string{"word", "word", "word", "word", "word", "word"}
	path, err := setupTestLoad(words)
	if err != nil {
		t.Fatalf(tests.ERR_FMT_STRING_AT_STRING, "setupTestLoad", err, path)
	}
	words, err = Load(path)
	if err != nil {
		t.Fatalf(tests.ERR_FMT_STRING_AT_STRING, "Load", err, path)
	}
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ERR_FMT_EXPECTED_GOT, "Load", expected, words)
	}
}

func TestFilterEmptyWords(t *testing.T) {
	words := []string{"", "", "word", " ", "  ", "", "", " word "}
	expected := []string{"word", " ", "  ", " word "}
	words = FilterEmptyWords(words)
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ERR_FMT_EXPECTED_GOT, "FilterEmptyWords", expected, words)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	words := []string{"word", "word", "word", "list", "list", "word", "list"}
	expected := []string{"word", "list"}
	words = RemoveDuplicates(words)
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ERR_FMT_EXPECTED_GOT, "RemoveDuplicates", expected, words)
	}
}

func TestFilterUTF8(t *testing.T) {
	words := []string{"word", "ẮẴÆƂƄḈḜ", "℥℗©ℌℹ", "∅⊆⊇∖", "▲◀►➣", "♔♛", "list"}
	expected := []string{"word", "list"}
	words = FilterUTF8(words)
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ERR_FMT_EXPECTED_GOT, "FilterUTF8", expected, words)
	}
}

func TestFromCSV(t *testing.T) {
	csv := "some, Large, LIST ,, of,simple,words,,as, expec ted,"
	expected := []string{"some", "Large", "LIST", "of", "simple", "words", "as", "expected"}
	words := FromCSV(csv)
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ERR_FMT_EXPECTED_GOT, "FromCSV", expected, words)
	}
}