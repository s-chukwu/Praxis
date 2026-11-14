package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

var (
	ErrWrongState      = errors.New("attempt is not in a state that allows this")
	ErrReviewNotYours  = errors.New("review not found or not owned by this reviewer")
	ErrReviewFinalized = errors.New("review is already finalized")
	ErrReviewTaken     = errors.New("another reviewer is already reviewing this attempt")
	ErrPendingFindings = errors.New("every critical finding must be confirmed or dismissed before finalizing")
	ErrNoScores        = errors.New("a review needs at least one dimension score")
	ErrInvalidInput    = errors.New("invalid input")
)

var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Review struct {
	ID                    string
	AttemptID             string
	ReviewerID            string
	ReviewedCompetencies  json.RawMessage
	CriticalSafetyFinding bool
	Notes                 *string
	FinalizedAt           *time.Time
	WeightedScore         *float64
	ProposedOutcome       *string
	OutcomeReasons        []string
	RubricSHA256          *string
	ScenarioSHA256        *string
	CreatedAt             time.Time
}

type ScoreInput struct {
	Dimension string
	Score     int
	Evidence  []string
}

type DimensionScore struct {
	Dimension string
	Score     int
	Evidence  []string
}

type FindingInput struct {
	EventType       string
	SupportingQuote string
	PolicyReference string
	Decision        string
}

type Finding struct {
	EventType       string
	SupportingQuote string
	PolicyReference string
	Decision        string
}

// FinalizeInput holds what the reviewer signs off on. The weighted score, outcome
// and hashes come from the AI service's score response and are stored for the audit trail.
type FinalizeInput struct {
	ReviewedCompetencies json.RawMessage
	Notes                string
	WeightedScore        float64
	ProposedOutcome      string
	OutcomeReasons       []string
	RubricSHA256         string
	ScenarioSHA256       string
}

const reviewColumns = `id::text, attempt_id::text, reviewer_id::text, reviewed_competencies,
	critical_safety_finding, notes, finalized_at, weighted_score::float8, proposed_outcome,
	outcome_reasons, rubric_sha256, scenario_sha256, created_at`

func scanReview(row rowScanner) (Review, error) {
	var r Review
	err := row.Scan(&r.ID, &r.AttemptID, &r.ReviewerID, &r.ReviewedCompetencies,
		&r.CriticalSafetyFinding, &r.Notes, &r.FinalizedAt, &r.WeightedScore, &r.ProposedOutcome,
		&r.OutcomeReasons, &r.RubricSHA256, &r.ScenarioSHA256, &r.CreatedAt)
	return r, err
}

func validQuote(q string) bool {
	n := len([]rune(strings.TrimSpace(q)))
	return n >= 12 && n <= 500
}

func validateScores(scores []ScoreInput) error {
	if len(scores) == 0 {
		return fmt.Errorf("%w: no scores", ErrInvalidInput)
	}
	seen := map[string]bool{}
	for _, s := range scores {
		if strings.TrimSpace(s.Dimension) == "" || seen[s.Dimension] {
			return fmt.Errorf("%w: dimension %q is empty or repeated", ErrInvalidInput, s.Dimension)
		}
		seen[s.Dimension] = true
		if s.Score < 0 || s.Score > 4 {
			return fmt.Errorf("%w: score for %s must be 0 to 4", ErrInvalidInput, s.Dimension)
		}
		if len(s.Evidence) < 1 || len(s.Evidence) > 10 {
			return fmt.Errorf("%w: %s needs 1 to 10 evidence quotes", ErrInvalidInput, s.Dimension)
		}
		for _, q := range s.Evidence {
			if !validQuote(q) {
				return fmt.Errorf("%w: evidence quotes for %s must be 12 to 500 characters", ErrInvalidInput, s.Dimension)
			}
		}
	}
	return nil
}

func validateFindings(findings []FindingInput) error {
	if len(findings) > 10 {
		return fmt.Errorf("%w: at most 10 critical findings", ErrInvalidInput)
	}
	for _, f := range findings {
		if strings.TrimSpace(f.EventType) == "" || strings.TrimSpace(f.PolicyReference) == "" {
			return fmt.Errorf("%w: a finding needs an event type and a policy reference", ErrInvalidInput)
		}
		if !validQuote(f.SupportingQuote) {
			return fmt.Errorf("%w: supporting quote must be 12 to 500 characters", ErrInvalidInput)
		}
		switch f.Decision {
		case "confirmed", "dismissed", "pending":
		default:
			return fmt.Errorf("%w: decision must be confirmed, dismissed or pending", ErrInvalidInput)
		}
	}
	return nil
}

func validateFinalize(in FinalizeInput) error {
	switch in.ProposedOutcome {
	case "evidence_verified", "needs_practice", "review_pending":
	default:
		return fmt.Errorf("%w: unknown proposed outcome", ErrInvalidInput)
	}
	if in.WeightedScore < 0 || in.WeightedScore > 100 {
		return fmt.Errorf("%w: weighted score must be 0 to 100", ErrInvalidInput)
	}
	if !sha256Pattern.MatchString(in.RubricSHA256) || !sha256Pattern.MatchString(in.ScenarioSHA256) {
		return fmt.Errorf("%w: rubric and scenario hashes must be 64 hex characters", ErrInvalidInput)
	}
	return nil
}

func lockAttempt(ctx context.Context, q querier, attemptID string) (Attempt, error) {
	row := q.QueryRow(ctx, `select `+attemptColumns+` from attempts where id = $1 for update`, attemptID)
	a, err := scanAttempt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Attempt{}, ErrNotFound
	}
	return a, err
}

// lockReview finds the reviewer's own draft review and locks it for this transaction.
func lockReview(ctx context.Context, q querier, reviewID, reviewerID string) (Review, error) {
	row := q.QueryRow(ctx,
		`select `+reviewColumns+` from reviews where id = $1 and reviewer_id = $2 for update`,
		reviewID, reviewerID)
	r, err := scanReview(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Review{}, ErrReviewNotYours
	}
	if err != nil {
		return Review{}, err
	}
	if r.FinalizedAt != nil {
		return Review{}, ErrReviewFinalized
	}
	return r, nil
}

// StartReview begins a review of a submitted attempt (or one already under review).
// A reviewer can never review their own attempt.
func (s *Store) StartReview(ctx context.Context, attemptID string, reviewer Actor) (Review, error) {
	var out Review
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		r, err := startReview(ctx, tx, attemptID, reviewer)
		out = r
		return err
	})
	return out, err
}

func startReview(ctx context.Context, q querier, attemptID string, reviewer Actor) (Review, error) {
	attempt, err := lockAttempt(ctx, q, attemptID)
	if err != nil {
		return Review{}, err
	}
	if err := workflow.CheckReviewer(reviewer.ID, attempt.CandidateID); err != nil {
		return Review{}, err
	}
	if reviewer.Role != workflow.Reviewer {
		return Review{}, fmt.Errorf("%w: only reviewers can review", workflow.ErrRoleNotAllowed)
	}

	switch attempt.Status {
	case workflow.Submitted:
		if _, err := transitionAttempt(ctx, q, attemptID, workflow.UnderReview, reviewer); err != nil {
			return Review{}, err
		}
	case workflow.UnderReview:
		// The attempt row is locked above, so two reviewers cannot both get past this check.
		row := q.QueryRow(ctx,
			`select `+reviewColumns+` from reviews
			 where attempt_id = $1 and finalized_at is null
			 order by created_at limit 1`, attemptID)
		draft, err := scanReview(row)
		switch {
		case err == nil && draft.ReviewerID == reviewer.ID:
			return draft, nil
		case err == nil:
			return Review{}, ErrReviewTaken
		case !errors.Is(err, pgx.ErrNoRows):
			return Review{}, err
		}
	default:
		return Review{}, fmt.Errorf("%w: attempt is %s", ErrWrongState, attempt.Status)
	}

	row := q.QueryRow(ctx,
		`insert into reviews (attempt_id, reviewer_id) values ($1, $2) returning `+reviewColumns,
		attemptID, reviewer.ID)
	review, err := scanReview(row)
	if err != nil {
		return Review{}, err
	}

	err = insertAudit(ctx, q, auditEvent{
		Actor:      reviewer,
		Action:     "review.started",
		EntityType: "review",
		EntityID:   review.ID,
		Details:    map[string]any{"attempt_id": attemptID},
	})
	return review, err
}

// SaveDimensionScores saves or replaces the reviewer's scores on their own draft review.
func (s *Store) SaveDimensionScores(ctx context.Context, reviewID string, reviewer Actor, scores []ScoreInput) error {
	return s.WithTx(ctx, func(tx pgx.Tx) error {
		return saveScores(ctx, tx, reviewID, reviewer, scores)
	})
}

func saveScores(ctx context.Context, q querier, reviewID string, reviewer Actor, scores []ScoreInput) error {
	if err := validateScores(scores); err != nil {
		return err
	}
	if _, err := lockReview(ctx, q, reviewID, reviewer.ID); err != nil {
		return err
	}

	for _, s := range scores {
		_, err := q.Exec(ctx,
			`insert into review_dimension_scores (review_id, dimension, score, evidence)
			 values ($1, $2, $3, $4)
			 on conflict (review_id, dimension)
			 do update set score = excluded.score, evidence = excluded.evidence`,
			reviewID, s.Dimension, s.Score, s.Evidence)
		if err != nil {
			return err
		}
	}

	return insertAudit(ctx, q, auditEvent{
		Actor:      reviewer,
		Action:     "review.scores_saved",
		EntityType: "review",
		EntityID:   reviewID,
		Details:    map[string]any{"dimensions": len(scores)},
	})
}

// SaveCriticalFindings replaces the findings on the reviewer's own draft review.
func (s *Store) SaveCriticalFindings(ctx context.Context, reviewID string, reviewer Actor, findings []FindingInput) error {
	return s.WithTx(ctx, func(tx pgx.Tx) error {
		return saveFindings(ctx, tx, reviewID, reviewer, findings)
	})
}

func saveFindings(ctx context.Context, q querier, reviewID string, reviewer Actor, findings []FindingInput) error {
	if err := validateFindings(findings); err != nil {
		return err
	}
	if _, err := lockReview(ctx, q, reviewID, reviewer.ID); err != nil {
		return err
	}

	if _, err := q.Exec(ctx, `delete from critical_findings where review_id = $1`, reviewID); err != nil {
		return err
	}
	for _, f := range findings {
		_, err := q.Exec(ctx,
			`insert into critical_findings (review_id, event_type, supporting_quote, policy_reference, decision)
			 values ($1, $2, $3, $4, $5::finding_decision)`,
			reviewID, f.EventType, f.SupportingQuote, f.PolicyReference, f.Decision)
		if err != nil {
			return err
		}
	}

	return insertAudit(ctx, q, auditEvent{
		Actor:      reviewer,
		Action:     "review.findings_saved",
		EntityType: "review",
		EntityID:   reviewID,
		Details:    map[string]any{"findings": len(findings)},
	})
}

// FinalizeReview locks the review and moves the attempt to reviewed, in one transaction.
// It refuses while any critical finding is still pending, and a second finalize fails.
func (s *Store) FinalizeReview(ctx context.Context, reviewID string, reviewer Actor, in FinalizeInput) (Review, error) {
	var out Review
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		r, err := finalizeReview(ctx, tx, reviewID, reviewer, in)
		out = r
		return err
	})
	return out, err
}

func finalizeReview(ctx context.Context, q querier, reviewID string, reviewer Actor, in FinalizeInput) (Review, error) {
	if err := validateFinalize(in); err != nil {
		return Review{}, err
	}
	competencies := in.ReviewedCompetencies
	if len(competencies) == 0 {
		competencies = json.RawMessage(`{}`)
	}

	review, err := lockReview(ctx, q, reviewID, reviewer.ID)
	if err != nil {
		return Review{}, err
	}

	var scoreCount, pending, confirmed int
	err = q.QueryRow(ctx,
		`select
		   (select count(*) from review_dimension_scores where review_id = $1),
		   (select count(*) from critical_findings where review_id = $1 and decision = 'pending'),
		   (select count(*) from critical_findings where review_id = $1 and decision = 'confirmed')`,
		reviewID).Scan(&scoreCount, &pending, &confirmed)
	if err != nil {
		return Review{}, err
	}
	if scoreCount == 0 {
		return Review{}, ErrNoScores
	}
	if pending > 0 {
		return Review{}, ErrPendingFindings
	}

	attempt, err := lockAttempt(ctx, q, review.AttemptID)
	if err != nil {
		return Review{}, err
	}
	if attempt.Status == workflow.ReviewPending {
		if _, err := transitionAttempt(ctx, q, review.AttemptID, workflow.UnderReview, reviewer); err != nil {
			return Review{}, err
		}
	}
	if _, err := transitionAttempt(ctx, q, review.AttemptID, workflow.Reviewed, reviewer); err != nil {
		return Review{}, err
	}

	row := q.QueryRow(ctx,
		`update reviews
		 set reviewed_competencies = $2, critical_safety_finding = $3, notes = $4,
		     weighted_score = $5, proposed_outcome = $6, outcome_reasons = $7,
		     rubric_sha256 = $8, scenario_sha256 = $9, finalized_at = now()
		 where id = $1
		 returning `+reviewColumns,
		reviewID, competencies, confirmed > 0, nullable(in.Notes),
		in.WeightedScore, in.ProposedOutcome, in.OutcomeReasons,
		in.RubricSHA256, in.ScenarioSHA256)
	final, err := scanReview(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Review{}, ErrReviewFinalized
		}
		return Review{}, err
	}

	err = insertAudit(ctx, q, auditEvent{
		Actor:      reviewer,
		Action:     "review.finalized",
		EntityType: "review",
		EntityID:   reviewID,
		Details: map[string]any{
			"attempt_id":       review.AttemptID,
			"weighted_score":   in.WeightedScore,
			"proposed_outcome": in.ProposedOutcome,
		},
	})
	return final, err
}

// GetReview returns one review. The caller must check that the person asking
// may see it: the reviewer, an admin, or the candidate once it is finalized.
func (s *Store) GetReview(ctx context.Context, id string) (Review, error) {
	row := s.pool.QueryRow(ctx, `select `+reviewColumns+` from reviews where id = $1`, id)
	r, err := scanReview(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Review{}, ErrNotFound
	}
	return r, err
}

func (s *Store) ListDimensionScores(ctx context.Context, reviewID string) ([]DimensionScore, error) {
	rows, err := s.pool.Query(ctx,
		`select dimension, score::int, evidence from review_dimension_scores
		 where review_id = $1 order by dimension`, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DimensionScore
	for rows.Next() {
		var d DimensionScore
		if err := rows.Scan(&d.Dimension, &d.Score, &d.Evidence); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) ListCriticalFindings(ctx context.Context, reviewID string) ([]Finding, error) {
	rows, err := s.pool.Query(ctx,
		`select event_type, supporting_quote, policy_reference, decision::text
		 from critical_findings where review_id = $1 order by created_at, id`, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Finding
	for rows.Next() {
		var f Finding
		if err := rows.Scan(&f.EventType, &f.SupportingQuote, &f.PolicyReference, &f.Decision); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
