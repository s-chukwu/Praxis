package store

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

// This test runs inside one transaction that is always rolled back,
// so it leaves nothing behind in the database (not even audit rows).
func TestAttemptLifecycle(t *testing.T) {
	s := connectForTest(t)
	ctx := context.Background()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var candidateID string
	err = tx.QueryRow(ctx, "select id::text from profiles where role = 'candidate' limit 1").Scan(&candidateID)
	if err != nil {
		t.Skipf("no candidate profile to test with: %v", err)
	}

	candidate := Actor{ID: candidateID, Role: workflow.Candidate}

	attempt, err := createAttempt(ctx, tx, NewAttempt{
		CandidateID:         candidateID,
		TargetTrack:         "operator",
		RoleStandardVersion: "1.0.0",
		ScenarioVersion:     "1.0.0",
		RubricVersion:       "0.1.0",
		PolicyVersion:       "1.0.0",
		ScenarioID:          "CS-01-A",
	}, candidate)
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	if attempt.Status != workflow.Assigned {
		t.Fatalf("new attempt should be assigned, got %s", attempt.Status)
	}
	if attempt.Stage != "baseline" {
		t.Fatalf("stage should default to baseline, got %s", attempt.Stage)
	}

	answers := json.RawMessage(`{"q1": "hello"}`)
	if _, err := saveAnswers(ctx, tx, attempt.ID, candidateID, answers); err != nil {
		t.Fatalf("save answers: %v", err)
	}

	if _, err := transitionAttempt(ctx, tx, attempt.ID, workflow.InProgress, candidate); err != nil {
		t.Fatalf("assigned to in_progress: %v", err)
	}
	submitted, err := transitionAttempt(ctx, tx, attempt.ID, workflow.Submitted, candidate)
	if err != nil {
		t.Fatalf("in_progress to submitted: %v", err)
	}
	if submitted.SubmittedAt == nil {
		t.Fatal("submitted_at should be set when an attempt is submitted")
	}

	// A candidate cannot start the review.
	_, err = transitionAttempt(ctx, tx, attempt.ID, workflow.UnderReview, candidate)
	if !errors.Is(err, workflow.ErrRoleNotAllowed) {
		t.Fatalf("candidate must not start a review, got %v", err)
	}

	// A reviewer cannot skip steps.
	reviewer := Actor{ID: candidateID, Role: workflow.Reviewer}
	_, err = transitionAttempt(ctx, tx, attempt.ID, workflow.Reviewed, reviewer)
	if !errors.Is(err, workflow.ErrIllegalTransition) {
		t.Fatalf("skipping steps must be refused, got %v", err)
	}

	// Answers are locked once the attempt is submitted.
	_, err = saveAnswers(ctx, tx, attempt.ID, candidateID, answers)
	if !errors.Is(err, ErrNotEditable) {
		t.Fatalf("answers must be locked after submit, got %v", err)
	}

	// One audit row for the creation and one for each of the two moves.
	var auditRows int
	err = tx.QueryRow(ctx, "select count(*) from audit_events where entity_id = $1::uuid", attempt.ID).Scan(&auditRows)
	if err != nil {
		t.Fatal(err)
	}
	if auditRows != 3 {
		t.Fatalf("expected 3 audit rows, got %d", auditRows)
	}
}

func TestTransitionUnknownAttempt(t *testing.T) {
	s := connectForTest(t)
	ctx := context.Background()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = transitionAttempt(ctx, tx, "00000000-0000-0000-0000-000000000000",
		workflow.InProgress, Actor{Role: workflow.System})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
