package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type NewPlanModule struct {
	Position            int
	CourseCode          string
	Title               string
	Covers              []string
	Reasons             json.RawMessage
	PreviouslyCompleted bool
}

// NewPlan holds a plan as the AI service returned it. Result is the full
// allocator result, kept as is for the audit trail.
type NewPlan struct {
	CandidateID          string
	ReviewID             string
	RequestID            string
	TargetTrack          string
	Status               string
	RoleStandardVersion  string
	PolicyVersion        string
	CurriculumVersion    string
	RoleStandardSHA256   string
	PolicySHA256         string
	CurriculumSHA256     string
	SafetyGateRequired   bool
	ReassessmentRequired bool
	Result               json.RawMessage
	Modules              []NewPlanModule
}

type PlanModule struct {
	Position            int
	CourseCode          string
	Title               string
	Covers              []string
	Reasons             json.RawMessage
	PreviouslyCompleted bool
	CompletedAt         *time.Time
	VerifiedAt          *time.Time
}

type Plan struct {
	ID                   string
	CandidateID          string
	ReviewID             string
	TargetTrack          string
	Status               string
	RoleStandardVersion  string
	PolicyVersion        string
	CurriculumVersion    string
	SafetyGateRequired   bool
	ReassessmentRequired bool
	CreatedAt            time.Time
	Modules              []PlanModule
}

const planColumns = `id::text, candidate_id::text, review_id::text, target_track, status,
	role_standard_version, policy_version, curriculum_version, safety_gate_required,
	reassessment_required, created_at`

func validatePlan(n NewPlan) error {
	required := []string{n.CandidateID, n.ReviewID, n.TargetTrack, n.RoleStandardVersion,
		n.PolicyVersion, n.CurriculumVersion}
	for _, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: a plan needs a candidate, review, track and versions", ErrInvalidInput)
		}
	}
	if n.Status != "plan_ready" && n.Status != "diagnostic_incomplete" {
		return fmt.Errorf("%w: unknown plan status", ErrInvalidInput)
	}
	if len(n.Result) == 0 {
		return fmt.Errorf("%w: the allocator result is required", ErrInvalidInput)
	}
	if len(n.Modules) > 50 {
		return fmt.Errorf("%w: too many modules", ErrInvalidInput)
	}
	seen := map[int]bool{}
	for _, m := range n.Modules {
		if m.Position < 1 || seen[m.Position] || strings.TrimSpace(m.CourseCode) == "" || strings.TrimSpace(m.Title) == "" {
			return fmt.Errorf("%w: every module needs a unique position, a course code and a title", ErrInvalidInput)
		}
		seen[m.Position] = true
	}
	return nil
}

// CreatePlan saves a learning plan and its modules in one transaction.
// It is idempotent: a plan for the same review and policy version is returned
// as it is, with created set to false.
func (s *Store) CreatePlan(ctx context.Context, n NewPlan, actor Actor) (Plan, bool, error) {
	var out Plan
	var created bool
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		p, c, err := createPlan(ctx, tx, n, actor)
		out, created = p, c
		return err
	})
	return out, created, err
}

func createPlan(ctx context.Context, q querier, n NewPlan, actor Actor) (Plan, bool, error) {
	if err := validatePlan(n); err != nil {
		return Plan{}, false, err
	}

	var planID string
	err := q.QueryRow(ctx,
		`insert into learning_plans (candidate_id, review_id, request_id, target_track, status,
		                             role_standard_version, policy_version, curriculum_version,
		                             role_standard_sha256, policy_sha256, curriculum_sha256,
		                             safety_gate_required, reassessment_required, result)
		 values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		 on conflict (review_id, policy_version) do nothing
		 returning id::text`,
		n.CandidateID, n.ReviewID, nullable(n.RequestID), n.TargetTrack, n.Status,
		n.RoleStandardVersion, n.PolicyVersion, n.CurriculumVersion,
		nullable(n.RoleStandardSHA256), nullable(n.PolicySHA256), nullable(n.CurriculumSHA256),
		n.SafetyGateRequired, n.ReassessmentRequired, n.Result).Scan(&planID)
	if errors.Is(err, pgx.ErrNoRows) {
		existing, err := getPlanByReview(ctx, q, n.ReviewID)
		return existing, false, err
	}
	if err != nil {
		return Plan{}, false, err
	}

	for _, m := range n.Modules {
		reasons := m.Reasons
		if len(reasons) == 0 {
			reasons = json.RawMessage(`[]`)
		}
		covers := m.Covers
		if covers == nil {
			covers = []string{}
		}
		_, err := q.Exec(ctx,
			`insert into learning_plan_modules (plan_id, position, course_code, title, covers, reasons, previously_completed)
			 values ($1, $2, $3, $4, $5, $6, $7)`,
			planID, m.Position, m.CourseCode, m.Title, covers, reasons, m.PreviouslyCompleted)
		if err != nil {
			return Plan{}, false, err
		}
	}

	err = insertAudit(ctx, q, auditEvent{
		Actor:      actor,
		Action:     "learning_plan.created",
		EntityType: "learning_plan",
		EntityID:   planID,
		Details:    map[string]any{"review_id": n.ReviewID, "status": n.Status, "modules": len(n.Modules)},
	})
	if err != nil {
		return Plan{}, false, err
	}

	plan, err := getPlanByReview(ctx, q, n.ReviewID)
	return plan, true, err
}

// GetPlanByReview returns the plan for a finalized review, or ErrNotFound.
// The caller must check that the person asking may see it.
func (s *Store) GetPlanByReview(ctx context.Context, reviewID string) (Plan, error) {
	return getPlanByReview(ctx, s.pool, reviewID)
}

func getPlanByReview(ctx context.Context, q querier, reviewID string) (Plan, error) {
	var p Plan
	err := q.QueryRow(ctx,
		`select `+planColumns+` from learning_plans where review_id = $1 order by created_at desc limit 1`,
		reviewID).Scan(&p.ID, &p.CandidateID, &p.ReviewID, &p.TargetTrack, &p.Status,
		&p.RoleStandardVersion, &p.PolicyVersion, &p.CurriculumVersion, &p.SafetyGateRequired,
		&p.ReassessmentRequired, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Plan{}, ErrNotFound
	}
	if err != nil {
		return Plan{}, err
	}

	rows, err := q.Query(ctx,
		`select position, course_code, title, covers, reasons, previously_completed, completed_at, verified_at
		 from learning_plan_modules where plan_id = $1 order by position`, p.ID)
	if err != nil {
		return Plan{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var m PlanModule
		if err := rows.Scan(&m.Position, &m.CourseCode, &m.Title, &m.Covers, &m.Reasons,
			&m.PreviouslyCompleted, &m.CompletedAt, &m.VerifiedAt); err != nil {
			return Plan{}, err
		}
		p.Modules = append(p.Modules, m)
	}
	return p, rows.Err()
}

// ListVerifiedModules returns the course codes a candidate has completed and had
// verified. They are sent to the AI service so it can skip satisfied prerequisites.
func (s *Store) ListVerifiedModules(ctx context.Context, candidateID string) ([]string, error) {
	return listVerifiedModules(ctx, s.pool, candidateID)
}

func listVerifiedModules(ctx context.Context, q querier, candidateID string) ([]string, error) {
	rows, err := q.Query(ctx,
		`select distinct m.course_code
		 from learning_plan_modules m
		 join learning_plans p on p.id = m.plan_id
		 where p.candidate_id = $1 and m.verified_at is not null
		 order by m.course_code`, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}
