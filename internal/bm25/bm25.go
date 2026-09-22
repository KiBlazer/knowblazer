package bm25

import (
	"math"
)

const (
	DefaultK1 = 1.2
	DefaultB  = 0.75
)

// Document represents an indexed unit with length and term frequencies.
type Document struct {
	Length int
	Freqs  map[string]int
}

// Corpus holds precomputed statistics across a set of documents.
type Corpus struct {
	TotalDocs int
	AvgDocLen float64
	DocFreq   map[string]int
	K1        float64
	B         float64
}

// NewCorpus creates a Corpus with default BM25 parameters k1=1.2, b=0.75.
func NewCorpus(docs []Document) *Corpus {
	c := &Corpus{
		TotalDocs: len(docs),
		DocFreq:   make(map[string]int),
		K1:        DefaultK1,
		B:         DefaultB,
	}
	if len(docs) == 0 {
		c.AvgDocLen = 1.0
		return c
	}

	totalLen := 0
	for _, doc := range docs {
		totalLen += doc.Length
		for term := range doc.Freqs {
			c.DocFreq[term]++
		}
	}

	c.AvgDocLen = float64(totalLen) / float64(len(docs))
	if c.AvgDocLen <= 0 {
		c.AvgDocLen = 1.0
	}
	return c
}

// Score computes the BM25 score for a document against a set of query terms.
func (c *Corpus) Score(doc Document, queryTerms []string) float64 {
	if c.TotalDocs == 0 || len(queryTerms) == 0 {
		return 0.0
	}

	score := 0.0
	for _, term := range queryTerms {
		tf := float64(doc.Freqs[term])
		if tf <= 0 {
			continue
		}

		df := float64(c.DocFreq[term])
		n := float64(c.TotalDocs)

		// Lucene/BM25 IDF: ln(1 + (N - n + 0.5) / (n + 0.5))
		idf := math.Log(1.0 + (n-df+0.5)/(df+0.5))

		// Length normalization
		lenNorm := 1.0 - c.B + c.B*(float64(doc.Length)/c.AvgDocLen)

		// BM25 term frequency saturation
		tfScore := (tf * (c.K1 + 1.0)) / (tf + c.K1*lenNorm)

		score += idf * tfScore
	}

	return score
}
