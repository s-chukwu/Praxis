package store

import (
	"errors"
	"testing"
)

func TestLowScoresCanRestOnWhatIsMissing(t *testing.T) {
	quote := []string{"I can see one completed payment."}
	ok := [][]ScoreInput{
		{{Dimension: "accuracy", Score: 0}},
		{{Dimension: "accuracy", Score: 1, Evidence: []string{}}},
		{{Dimension: "accuracy", Score: 3, Evidence: quote}},
		{{Dimension: "accuracy", Score: 0, Evidence: quote}},
	}
	for _, s := range ok {
		if err := validateScores(s); err != nil {
			t.Errorf("%+v should be valid, got %v", s, err)
		}
	}
	if err := validateScores([]ScoreInput{{Dimension: "accuracy", Score: 2}}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("a score of 2 with no quote must be refused, got %v", err)
	}
}
