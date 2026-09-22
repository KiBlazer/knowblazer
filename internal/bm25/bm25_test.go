package bm25

import (
	"testing"
)

func TestBM25ScoringFavorsCompactDocOverDilutedDoc(t *testing.T) {
	docs := []Document{
		{
			Length: 20,
			Freqs: map[string]int{
				"grpc":     2,
				"timeout":  2,
				"deadline": 2,
			},
		},
		{
			Length: 3000,
			Freqs: map[string]int{
				"grpc":     5,
				"timeout":  5,
				"deadline": 5,
			},
		},
	}

	corpus := NewCorpus(docs)
	queryTerms := []string{"grpc", "timeout", "deadline"}

	scoreFocused := corpus.Score(docs[0], queryTerms)
	scoreDiluted := corpus.Score(docs[1], queryTerms)

	if scoreFocused <= scoreDiluted {
		t.Fatalf("expected focused note to score higher than diluted note: focused=%v, diluted=%v", scoreFocused, scoreDiluted)
	}
}

func TestBM25IDFCalculation(t *testing.T) {
	docs := []Document{
		{
			Length: 100,
			Freqs: map[string]int{
				"common": 2,
				"rare":   1,
			},
		},
		{
			Length: 100,
			Freqs: map[string]int{
				"common": 2,
			},
		},
		{
			Length: 100,
			Freqs: map[string]int{
				"common": 2,
			},
		},
	}

	corpus := NewCorpus(docs)

	// Rare term should yield higher IDF and thus higher score than common term
	scoreRare := corpus.Score(docs[0], []string{"rare"})
	scoreCommon := corpus.Score(docs[0], []string{"common"})

	if scoreRare <= scoreCommon {
		t.Fatalf("expected rare term to score higher due to IDF: rare=%v, common=%v", scoreRare, scoreCommon)
	}
}
