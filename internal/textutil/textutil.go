package textutil

import (
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

// Tokenize extracts lowercase alphanumeric terms and CJK unigram/bigram tokens.
func Tokenize(text string) []string {
	text = strings.ToLower(text)
	var out []string
	seen := map[string]bool{}

	var latinBuf []rune
	var cjkBuf []rune

	flushLatin := func() {
		if len(latinBuf) >= 2 {
			word := string(latinBuf)
			if !seen[word] {
				seen[word] = true
				out = append(out, word)
			}
		}
		latinBuf = latinBuf[:0]
	}

	flushCJK := func() {
		n := len(cjkBuf)
		if n == 0 {
			return
		}
		if n >= 2 {
			full := string(cjkBuf)
			if !seen[full] {
				seen[full] = true
				out = append(out, full)
			}
			for i := 0; i < n-1; i++ {
				bigram := string(cjkBuf[i : i+2])
				if !seen[bigram] {
					seen[bigram] = true
					out = append(out, bigram)
				}
			}
		}
		for i := 0; i < n; i++ {
			unigram := string(cjkBuf[i])
			if !seen[unigram] {
				seen[unigram] = true
				out = append(out, unigram)
			}
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

	return out
}
