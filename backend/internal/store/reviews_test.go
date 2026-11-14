package store

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

// Runs inside one transaction that is always rolled back, so nothing is left behind.
func TestReviewFlow(t *testing.T) {
	s := connectForTest(t)
	ctx := context.Background()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var candidateID, reviewerID string
	err = tx.QueryRow(ctx, "select id::text from profiles where role = 'candidate' limit 1").Scan(&candidateID)
	if err != nil {
		t.Skipf("no candidate profile to test with: %v", err)
	}
	err = tx.QueryRow(ctx, "select id::text from profiles where role = 'reviewer' limit 1").Scan(&reviewerID)
	if err != nil {
		t.Skipf("no reviewer profile to test with: %v", err)
	}

	candidate := Actor{ID: candidateID, Role: workflow.Candidate}
	reviewer := Actor{ID: reviewerID, Role: workflow.Reviewer}

	attempt, err := createAttempt(ctx, tx, NewAttempt{
		CandidateID:         candidateID,
		TargetTrack:         "operator",
		RoleStandardVersion: "1.0.0",
		ScenarioVersion:     "1.1.0",
		RubricVersion:       "0.2.0",
		PolicyVersion:       "1.0.0",
		ScenarioID:          "CS-01-A",
	}, candidate)
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	for _, to := range []workflow.State{workflow.InProgress, workflow.Submitted} {
		if _, err := transitionAttempt(ctx, tx, attempt.ID, to, candidate); err != nil {
			t.Fatalf("move to %s: %v", to, err)
		}
	}

	// A person can never review their own attempt.
	_, err = startReview(ctx, tx, attempt.ID, Actor{ID: candidateID, Role: workflow.Reviewer})
	if !errors.Is(err, workflow.ErrSelfReview) {
		t.Fatalf("self review must be refused, got %v", err)
	}

	review, err := startReview(ctx, tx, attempt.ID, reviewer)
	if err != nil {
		t.Fatalf("start review: %v", err)
	}

	again, err := startReview(ctx, tx, attempt.ID, reviewer)
	if err != nil || again.ID != review.ID {
		t.Fatalf("starting again should return the same draft, got %v %v", again.ID, err)
	}
	_, err = startReview(ctx, tx, attempt.ID, Actor{ID: "00000000-0000-4000-8000-000000000001", Role: workflow.Reviewer})
	if !errors.Is(err, ErrReviewTaken) {
		t.Fatalf("a second reviewer must be refused, got %v", err)
	}

	var status string
	if err := tx.QueryRow(ctx, "select status::text from attempts where id = $1", attempt.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "under_review" {
		t.Fatalf("attempt should be under_review, got %s", status)
	}

	quote := "I can help you reset that password now."

	// Invalid input is refused before it touches the database.
	err = saveScores(ctx, tx, review.ID, reviewer, []ScoreInput{{Dimension: "accuracy", Score: 5, Evidence: []string{quote}}})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("score 5 must be refused, got %v", err)
	}
	err = saveScores(ctx, tx, review.ID, reviewer, []ScoreInput{{Dimension: "accuracy", Score: 3, Evidence: []string{"too short"}}})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("a short quote must be refused, got %v", err)
	}

	scores := []ScoreInput{
		{Dimension: "accuracy", Score: 3, Evidence: []string{quote}},
		{Dimension: "judgment", Score: 3, Evidence: []string{quote}},
		{Dimension: "safety", Score: 4, Evidence: []string{quote}},
		{Dimension: "communication", Score: 3, Evidence: []string{quote}},
	}
	if err := saveScores(ctx, tx, review.ID, reviewer, scores); err != nil {
		t.Fatalf("save scores: %v", err)
	}

	pending := []FindingInput{{
		EventType:       "unauthorized_action",
		SupportingQuote: quote,
		PolicyReference: "POL-1",
		Decision:        "pending",
	}}
	if err := saveFindings(ctx, tx, review.ID, reviewer, pending); err != nil {
		t.Fatalf("save findings: %v", err)
	}

	in := FinalizeInput{
		Notes:           "Reviewed against the rubric.",
		WeightedScore:   82.5,
		ProposedOutcome: "evidence_verified",
		OutcomeReasons:  []string{"meets_pilot_threshold"},
		RubricSHA256:    strings.Repeat("a", 64),
		ScenarioSHA256:  strings.Repeat("b", 64),
	}

	// A pending finding blocks finalizing.
	if _, err := finalizeReview(ctx, tx, review.ID, reviewer, in); !errors.Is(err, ErrPendingFindings) {
		t.Fatalf("pending findings must block finalizing, got %v", err)
	}

	confirmed := []FindingInput{{
		EventType:       "unauthorized_action",
		SupportingQuote: quote,
		PolicyReference: "POL-1",
		Decision:        "confirmed",
	}}
	if err := saveFindings(ctx, tx, review.ID, reviewer, confirmed); err != nil {
		t.Fatalf("adjudicate finding: %v", err)
	}

	final, err := finalizeReview(ctx, tx, review.ID, reviewer, in)
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}
	if final.FinalizedAt == nil || !final.CriticalSafetyFinding {
		t.Fatalf("review should be finalized with the safety finding set: %+v", final)
	}

	if err := tx.QueryRow(ctx, "select status::text from attempts where id = $1", attempt.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "reviewed" {
		t.Fatalf("attempt should be reviewed, got %s", status)
	}

	// Once finalized, nothing can change.
	if _, err := finalizeReview(ctx, tx, review.ID, reviewer, in); !errors.Is(err, ErrReviewFinalized) {
		t.Fatalf("second finalize must be refused, got %v", err)
	}
	if err := saveScores(ctx, tx, review.ID, reviewer, scores); !errors.Is(err, ErrReviewFinalized) {
		t.Fatalf("scores must be locked after finalizing, got %v", err)
	}
}
