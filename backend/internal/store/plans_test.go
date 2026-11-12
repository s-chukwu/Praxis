package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

// Runs inside one transaction that is always rolled back, so nothing is left behind.
func TestPlanLifecycle(t *testing.T) {
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

	// A finalized review is needed first, because a plan belongs to one.
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
	review, err := startReview(ctx, tx, attempt.ID, reviewer)
	if err != nil {
		t.Fatalf("start review: %v", err)
	}
	quote := "I can help you reset that password now."
	scores := []ScoreInput{{Dimension: "accuracy", Score: 3, Evidence: []string{quote}}}
	if err := saveScores(ctx, tx, review.ID, reviewer, scores); err != nil {
		t.Fatalf("save scores: %v", err)
	}
	_, err = finalizeReview(ctx, tx, review.ID, reviewer, FinalizeInput{
		ReviewedCompetencies: json.RawMessage(`{"C1":3,"C4":2}`),
		WeightedScore:        70,
		ProposedOutcome:      "needs_practice",
		OutcomeReasons:       []string{"below_pilot_threshold"},
		RubricSHA256:         strings.Repeat("a", 64),
		ScenarioSHA256:       strings.Repeat("b", 64),
	})
	if err != nil {
		t.Fatalf("finalize review: %v", err)
	}

	plan := NewPlan{
		CandidateID:          candidateID,
		ReviewID:             review.ID,
		RequestID:            "44444444-4444-4444-4444-444444444444",
		TargetTrack:          "operator",
		Status:               "plan_ready",
		RoleStandardVersion:  "1.0.0",
		PolicyVersion:        "1.0.0",
		CurriculumVersion:    "1.0.0",
		RoleStandardSHA256:   strings.Repeat("c", 64),
		PolicySHA256:         strings.Repeat("d", 64),
		CurriculumSHA256:     strings.Repeat("e", 64),
		SafetyGateRequired:   true,
		ReassessmentRequired: true,
		Result:               json.RawMessage(`{"status":"plan_ready"}`),
		Modules: []NewPlanModule{
			{Position: 2, CourseCode: "M-SAFE", Title: "Safety and privacy", Covers: []string{"C4"}},
			{Position: 1, CourseCode: "M-ROLE", Title: "Role orientation", Covers: []string{"C1", "C3"}, Reasons: json.RawMessage(`[{"code":"role_orientation"}]`)},
		},
	}

	created, wasCreated, err := createPlan(ctx, tx, plan, reviewer)
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if !wasCreated || len(created.Modules) != 2 {
		t.Fatalf("expected a new plan with 2 modules, got created=%v modules=%d", wasCreated, len(created.Modules))
	}
	if created.Modules[0].CourseCode != "M-ROLE" || created.Modules[1].CourseCode != "M-SAFE" {
		t.Fatalf("modules must come back in position order, got %+v", created.Modules)
	}

	// The same review and policy version gives back the same plan.
	again, wasCreated, err := createPlan(ctx, tx, plan, reviewer)
	if err != nil {
		t.Fatalf("create plan again: %v", err)
	}
	if wasCreated || again.ID != created.ID {
		t.Fatalf("a repeated request must return the existing plan, got created=%v", wasCreated)
	}

	// Bad input is refused.
	bad := plan
	bad.Modules = []NewPlanModule{{Position: 1, CourseCode: " ", Title: "x"}}
	if _, _, err := createPlan(ctx, tx, bad, reviewer); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("a module without a course code must be refused, got %v", err)
	}

	// Verified modules are reported for the AI service.
	before, err := listVerifiedModules(ctx, tx, candidateID)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != 0 {
		t.Fatalf("no module is verified yet, got %v", before)
	}
	_, err = tx.Exec(ctx, `update learning_plan_modules set verified_at = now()
		where plan_id = $1 and course_code = 'M-ROLE'`, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	after, err := listVerifiedModules(ctx, tx, candidateID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 1 || after[0] != "M-ROLE" {
		t.Fatalf("expected [M-ROLE], got %v", after)
	}

	// The creation was written to the audit log.
	var auditRows int
	err = tx.QueryRow(ctx, "select count(*) from audit_events where entity_id = $1::uuid and action = 'learning_plan.created'",
		created.ID).Scan(&auditRows)
	if err != nil {
		t.Fatal(err)
	}
	if auditRows != 1 {
		t.Fatalf("expected 1 audit row for the plan, got %d", auditRows)
	}
}
