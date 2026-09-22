package textutil

import (
	"sort"
	"strings"
	"unicode"
)

// IsCJK reports whether the rune is a CJK ideograph or phonetic character.
func IsCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hangul, r)
}

// Tokenize extracts unique lowercase alphanumeric terms and CJK unigram/bigram tokens.
func Tokenize(text string) []string {
	counts, _ := TermFrequencies(text)
	out := make([]string, 0, len(counts))
	for term := range counts {
		out = append(out, term)
	}
	sort.Strings(out)
	return out
}

// TermFrequencies extracts term occurrences with their frequencies and total token count (for BM25).
func TermFrequencies(text string) (map[string]int, int) {
	text = strings.ToLower(text)
	counts := make(map[string]int)
	totalTokens := 0

	var latinBuf []rune
	var cjkBuf []rune

	flushLatin := func() {
		if len(latinBuf) >= 2 {
			word := string(latinBuf)
			counts[word]++
			totalTokens++
		}
		latinBuf = latinBuf[:0]
	}

	flushCJK := func() {
		n := len(cjkBuf)
		if n == 0 {
			return
		}
		totalTokens += n
		if n >= 2 {
			full := string(cjkBuf)
			counts[full]++
			for i := 0; i < n-1; i++ {
				bigram := string(cjkBuf[i : i+2])
				counts[bigram]++
			}
		}
		for i := 0; i < n; i++ {
			unigram := string(cjkBuf[i])
			counts[unigram]++
		}
		cjkBuf = cjkBuf[:0]
	}

	for _, r := range text {
		if IsCJK(r) {
			flushLatin()
			cjkBuf = append(cjkBuf, r)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			flushCJK()
			latinBuf = append(latinBuf, r)
		} else {
			flushLatin()
			flushCJK()
		}
	}
	flushLatin()
	flushCJK()

	if totalTokens == 0 {
		totalTokens = 1
	}
	return counts, totalTokens
}

