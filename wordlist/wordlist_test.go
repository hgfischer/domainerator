package wordlist

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/hgfischer/domainerator/tests"
)

func TestTrimWords(t *testing.T) {
	words := []string{" word ", "word  ", "    word", "wor d", "   wo rd ", "   wo    rd "}
	expected := []string{"word", "word", "word", "word", "word", "word"}
	words = TrimWords(words)
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ErrFmtExpectedGot, "TrimWords", expected, words)
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
		t.Fatalf(tests.ErrFmtStringAtString, "setupTestLoad", err, path)
	}
	words, err = Load(path)
	if err != nil {
		t.Fatalf(tests.ErrFmtStringAtString, "Load", err, path)
	}
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ErrFmtExpectedGot, "Load", expected, words)
	}
}

func TestFilterEmptyWords(t *testing.T) {
	words := []string{"", "", "word", " ", "  ", "", "", " word "}
	expected := []string{"word", " ", "  ", " word "}
	words = FilterEmptyWords(words)
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ErrFmtExpectedGot, "FilterEmptyWords", expected, words)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	words := []string{"word", "word", "word", "list", "list", "word", "list"}
	expected := []string{"word", "list"}
	words = RemoveDuplicates(words)
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ErrFmtExpectedGot, "RemoveDuplicates", expected, words)
	}
}

func TestFilterUTF8(t *testing.T) {
	words := []string{"word", "ẮẴÆƂƄḈḜ", "℥℗©ℌℹ", "∅⊆⊇∖", "▲◀►➣", "♔♛", "list"}
	expected := []string{"word", "list"}
	words = FilterUTF8(words)
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ErrFmtExpectedGot, "FilterUTF8", expected, words)
	}
}

func TestFromCSV(t *testing.T) {
	csv := "some, Large, LIST ,, of,simple,words,,as, expec ted,"
	expected := []string{"some", "Large", "LIST", "of", "simple", "words", "as", "expected"}
	words := FromCSV(csv)
	if !reflect.DeepEqual(words, expected) {
		t.Errorf(tests.ErrFmtExpectedGot, "FromCSV", expected, words)
	}
}
