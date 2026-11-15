package store

import (
	"context"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

func TestBaselineCanRestartAfterVoid(t *testing.T) {
	s := connectForTest(t)
	ctx := context.Background()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var candidateID, adminID string
	err = tx.QueryRow(ctx,
		`select id::text from profiles p
		 where p.role = 'candidate'
		   and not exists (select 1 from attempts a where a.candidate_id = p.id and a.stage = 'baseline')
		 limit 1`).Scan(&candidateID)
	if err != nil {
		t.Skipf("no candidate without a baseline attempt to test with: %v", err)
	}
	err = tx.QueryRow(ctx, "select id::text from profiles where role = 'admin' limit 1").Scan(&adminID)
	if err != nil {
		t.Skipf("no admin profile to test with: %v", err)
	}

	candidate := Actor{ID: candidateID, Role: workflow.Candidate}
	admin := Actor{ID: adminID, Role: workflow.Role("admin")}
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
	if err != nil || !created {
		t.Fatalf("first start: created=%v err=%v", created, err)
	}
	for _, to := range []workflow.State{workflow.InProgress, workflow.Submitted} {
		if _, err := transitionAttempt(ctx, tx, first.ID, to, candidate); err != nil {
			t.Fatalf("move to %s: %v", to, err)
		}
	}
	if _, err := transitionAttempt(ctx, tx, first.ID, workflow.Voided, admin); err != nil {
		t.Fatalf("admin void: %v", err)
	}

	// After the void, the candidate gets a brand new attempt.
	second, created, err := createOrGetBaselineAttempt(ctx, tx, n, candidate)
	if err != nil {
		t.Fatalf("start after void: %v", err)
	}
	if !created || second.ID == first.ID || second.Status != workflow.Assigned {
		t.Fatalf("expected a new assigned attempt, got created=%v %+v", created, second)
	}

	// Starting again returns that new attempt, not another one.
	again, created, err := createOrGetBaselineAttempt(ctx, tx, n, candidate)
	if err != nil || created || again.ID != second.ID {
		t.Fatalf("expected the same new attempt, got created=%v id=%s err=%v", created, again.ID, err)
	}

	var total, active int
	err = tx.QueryRow(ctx,
		`select count(*), count(*) filter (where status <> 'voided')
		 from attempts where candidate_id = $1 and stage = 'baseline'`, candidateID).Scan(&total, &active)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || active != 1 {
		t.Fatalf("expected 2 attempts with 1 active, got %d and %d", total, active)
	}
}