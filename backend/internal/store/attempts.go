package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrNotEditable = errors.New("attempt not found or no longer editable")
)

// querier is satisfied by both the connection pool and a transaction,
// so the same function can run on either (tests use a transaction they roll back).
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Attempt struct {
	ID                  string
	CandidateID         string
	TargetTrack         string
	RoleStandardVersion string

	ScenarioVersion     string
	RubricVersion       string
	PolicyVersion       string
	ScenarioID          *string
	Stage               string
	Status              workflow.State
	Answers             json.RawMessage
	SubmittedAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type NewAttempt struct {
	CandidateID         string
	TargetTrack         string
	RoleStandardVersion string
	ScenarioVersion     string
	RubricVersion       string
	PolicyVersion       string
	ScenarioID          string
	Stage               string
}

const attemptColumns = `id::text, candidate_id::text, target_track, role_standard_version,
	scenario_version, rubric_version, policy_version, scenario_id, stage, status::text,
	answers, submitted_at, created_at, updated_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAttempt(row rowScanner) (Attempt, error) {

	var a Attempt
	var status string
	err := row.Scan(&a.ID, &a.CandidateID, &a.TargetTrack, &a.RoleStandardVersion,
		&a.ScenarioVersion, &a.RubricVersion, &a.PolicyVersion, &a.ScenarioID, &a.Stage,
		&status, &a.Answers, &a.SubmittedAt, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return Attempt{}, err
	}
	a.Status = workflow.State(status)
	return a, nil
}

func collectAttempts(rows pgx.Rows) ([]Attempt, error) {
	defer rows.Close()

	var out []Attempt
	for rows.Next() {
		a, err := scanAttempt(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// CreateAttempt starts a new attempt in the assigned state and writes an audit event.
func (s *Store) CreateAttempt(ctx context.Context, n NewAttempt, actor Actor) (Attempt, error) {
	var out Attempt
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		a, err := createAttempt(ctx, tx, n, actor)
		out = a

		return err
	})
	return out, err
}

func createAttempt(ctx context.Context, q querier, n NewAttempt, actor Actor) (Attempt, error) {
	stage := n.Stage
	if stage == "" {
		stage = "baseline"
	}

	row := q.QueryRow(ctx,
		`insert into attempts (candidate_id, target_track, role_standard_version, scenario_version,
		                       rubric_version, policy_version, scenario_id, stage)
		 values ($1, $2, $3, $4, $5, $6, $7, $8)
		 returning `+attemptColumns,
		n.CandidateID, n.TargetTrack, n.RoleStandardVersion, n.ScenarioVersion,
		n.RubricVersion, n.PolicyVersion, nullable(n.ScenarioID), stage)

	a, err := scanAttempt(row)
	if err != nil {
		return Attempt{}, err
	}

	err = insertAudit(ctx, q, auditEvent{
		Actor:      actor,
		Action:     "attempt.created",
		EntityType: "attempt",
		EntityID:   a.ID,
		Details:    map[string]any{"scenario_id": n.ScenarioID, "stage": stage},
	})
	return a, err

}

// CreateOrGetBaselineAttempt gives a candidate their one baseline attempt.
// If they already have one, it is returned untouched (answers and progress kept)
// and created is false. If not, a new one is made and created is true.
// The database index attempts_one_baseline_per_candidate guarantees that two
// requests at the same moment can never create two attempts.
func (s *Store) CreateOrGetBaselineAttempt(ctx context.Context, n NewAttempt, actor Actor) (Attempt, bool, error) {
	var out Attempt
	var created bool
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		a, c, err := createOrGetBaselineAttempt(ctx, tx, n, actor)
		out, created = a, c
		return err
	})
	return out, created, err
}

func createOrGetBaselineAttempt(ctx context.Context, q querier, n NewAttempt, actor Actor) (Attempt, bool, error) {
	n.Stage = "baseline"

	// Try to insert. If this candidate already has a baseline attempt, the insert
	// does nothing and returns no row.
	row := q.QueryRow(ctx,
		`insert into attempts (candidate_id, target_track, role_standard_version, scenario_version,
		                       rubric_version, policy_version, scenario_id, stage)
		 values ($1, $2, $3, $4, $5, $6, $7, 'baseline')
		 on conflict (candidate_id) where stage = 'baseline' do nothing
		 returning `+attemptColumns,
		n.CandidateID, n.TargetTrack, n.RoleStandardVersion, n.ScenarioVersion,
		n.RubricVersion, n.PolicyVersion, nullable(n.ScenarioID))


	a, err := scanAttempt(row)
	if err == nil {
		// A new attempt was created, so record it.
		err = insertAudit(ctx, q, auditEvent{
			Actor:      actor,
			Action:     "attempt.created",
			EntityType: "attempt",
			EntityID:   a.ID,
			Details:    map[string]any{"scenario_id": n.ScenarioID, "stage": "baseline"},
		})
		if err != nil {
			return Attempt{}, false, err
		}
		return a, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Attempt{}, false, err
	}

	// Nothing was inserted, so one already exists. Return it as it is.
	row = q.QueryRow(ctx,
		`select `+attemptColumns+` from attempts
		 where candidate_id = $1 and stage = 'baseline'`,
		n.CandidateID)
	existing, err := scanAttempt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Attempt{}, false, ErrNotFound
	}
	if err != nil {
		return Attempt{}, false, err
	}
	return existing, false, nil

}

// GetAttempt returns one attempt. The caller must check that the person
// asking is allowed to see it (owner, reviewer or admin) before returning it.
func (s *Store) GetAttempt(ctx context.Context, id string) (Attempt, error) {
	row := s.pool.QueryRow(ctx, `select `+attemptColumns+` from attempts where id = $1`, id)
	a, err := scanAttempt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Attempt{}, ErrNotFound
	}
	return a, err
}

// ListCandidateAttempts returns a candidate's own attempts, newest first.
func (s *Store) ListCandidateAttempts(ctx context.Context, candidateID string) ([]Attempt, error) {
	rows, err := s.pool.Query(ctx,
		`select `+attemptColumns+` from attempts where candidate_id = $1 order by created_at desc`,
		candidateID)
	if err != nil {
		return nil, err
	}
	return collectAttempts(rows)
}

// ListReviewQueue returns attempts waiting for or in review, oldest submission first.
func (s *Store) ListReviewQueue(ctx context.Context) ([]Attempt, error) {
	rows, err := s.pool.Query(ctx,
		`select `+attemptColumns+` from attempts
		 where status in ('submitted', 'under_review', 'review_pending')
		 order by submitted_at nulls last, created_at`)
	if err != nil {
		return nil, err

	}
	return collectAttempts(rows)
}

// SaveAnswers stores a candidate's answers while the attempt is still open
// (assigned or in progress) and belongs to that candidate.
func (s *Store) SaveAnswers(ctx context.Context, attemptID, candidateID string, answers json.RawMessage) (Attempt, error) {
	return saveAnswers(ctx, s.pool, attemptID, candidateID, answers)
}

func saveAnswers(ctx context.Context, q querier, attemptID, candidateID string, answers json.RawMessage) (Attempt, error) {
	row := q.QueryRow(ctx,
		`update attempts set answers = $3
		 where id = $1 and candidate_id = $2 and status in ('assigned', 'in_progress')
		 returning `+attemptColumns,
		attemptID, candidateID, answers)

	a, err := scanAttempt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Attempt{}, ErrNotEditable
	}
	return a, err
}

// TransitionAttempt moves an attempt to a new state. The state and role check
// runs inside the same transaction as the update, with the row locked, so two
// requests cannot both pass the check. The service role skips the database's
// own role checks, so this Go check is the one that matters for roles.
func (s *Store) TransitionAttempt(ctx context.Context, attemptID string, to workflow.State, actor Actor) (Attempt, error) {
	var out Attempt
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		a, err := transitionAttempt(ctx, tx, attemptID, to, actor)

		out = a
		return err
	})
	return out, err
}

func transitionAttempt(ctx context.Context, q querier, attemptID string, to workflow.State, actor Actor) (Attempt, error) {
	row := q.QueryRow(ctx, `select `+attemptColumns+` from attempts where id = $1 for update`, attemptID)
	current, err := scanAttempt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Attempt{}, ErrNotFound
	}
	if err != nil {
		return Attempt{}, err
	}

	if err := workflow.CheckTransition(current.Status, to, actor.Role); err != nil {
		return Attempt{}, err
	}

	row = q.QueryRow(ctx,
		`update attempts
		 set status = $2::attempt_status,
		     submitted_at = case when $2::attempt_status = 'submitted' then now() else submitted_at end
		 where id = $1
		 returning `+attemptColumns,
		attemptID, string(to))
	updated, err := scanAttempt(row)
	if err != nil {
		return Attempt{}, err
	}


	err = insertAudit(ctx, q, auditEvent{
		Actor:      actor,
		Action:     "attempt.transition",
		EntityType: "attempt",
		EntityID:   attemptID,
		Details:    map[string]any{"from": string(current.Status), "to": string(to)},
	})
	return updated, err
}