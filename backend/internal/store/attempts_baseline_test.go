package store

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

// Runs inside one transaction that is always rolled back, so nothing is left behind.
// It only runs when DATABASE_URL is set in the terminal.
func TestBaselineAttemptIsCreatedOnce(t *testing.T) {
	s := connectForTest(t)
	ctx := context.Background()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// A candidate who does not have a baseline attempt yet.
	var candidateID string
	err = tx.QueryRow(ctx,
		`select id::text from profiles p
		 where p.role = 'candidate'
		   and not exists (select 1 from attempts a where a.candidate_id = p.id and a.stage = 'baseline')
		 limit 1`).Scan(&candidateID)
	if err != nil {
		t.Skipf("no candidate without a baseline attempt to test with: %v", err)

	}

	candidate := Actor{ID: candidateID, Role: workflow.Candidate}
	n := NewAttempt{
		CandidateID:         candidateID,
		TargetTrack:         "operator",
		RoleStandardVersion: "1.0.0",
		ScenarioVersion:     "1.1.0",
		RubricVersion:       "0.2.0",
		PolicyVersion:       "1.0.0",
		ScenarioID:          "CS-01-A",
	}

	first, created, err := createOrGetBaselineAttempt(ctx, tx, n, candidate)
	if err != nil {
		t.Fatalf("first start: %v", err)
	}
	if !created {
		t.Fatalf("the first start must create the attempt")
	}
	if first.Stage != "baseline" || first.Status != workflow.Assigned {
		t.Fatalf("unexpected attempt: %+v", first)
	}

	// The candidate saves some work.
	if _, err := saveAnswers(ctx, tx, first.ID, candidateID, json.RawMessage(`{"q1":"hello"}`)); err != nil {
		t.Fatalf("save answers: %v", err)
	}

	// Starting again returns the same attempt with the answers kept.
	again, created, err := createOrGetBaselineAttempt(ctx, tx, n, candidate)
	if err != nil {

		t.Fatalf("second start: %v", err)
	}
	if created {
		t.Fatalf("the second start must not create a new attempt")
	}
	if again.ID != first.ID {
		t.Fatalf("expected the same attempt, got %s and %s", first.ID, again.ID)
	}
	if !strings.Contains(string(again.Answers), "hello") {
		t.Fatalf("saved answers must be kept, got %s", again.Answers)
	}

	// Still only one baseline attempt for this candidate.
	var count int
	err = tx.QueryRow(ctx,
		`select count(*) from attempts where candidate_id = $1 and stage = 'baseline'`,
		candidateID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one baseline attempt, got %d", count)
	}
}