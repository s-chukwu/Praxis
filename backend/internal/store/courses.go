package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// CompleteCourse marks one course of the candidate's latest plan as done. Courses unlock in plan
// order, so a course is refused while an earlier one is still open. Completing twice is a no-op.
func (s *Store) CompleteCourse(ctx context.Context, candidateID, courseCode string, actor Actor) (time.Time, error) {
	var done time.Time
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		var planID string
		err := tx.QueryRow(ctx,
			`select id::text from learning_plans where candidate_id = $1 order by created_at desc limit 1`, candidateID).Scan(&planID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}

		var position int
		var completed *time.Time
		err = tx.QueryRow(ctx,
			`select position, completed_at from learning_plan_modules
			 where plan_id = $1 and course_code = $2 for update`, planID, courseCode).Scan(&position, &completed)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if completed != nil {
			done = *completed
			return nil
		}

		var open int
		if err := tx.QueryRow(ctx,
			`select count(*) from learning_plan_modules where plan_id = $1 and position < $2 and completed_at is null`,
			planID, position).Scan(&open); err != nil {
			return err
		}
		if open > 0 {
			return fmt.Errorf("%w: finish the earlier courses first", ErrWrongState)
		}

		if err := tx.QueryRow(ctx,
			`update learning_plan_modules set completed_at = now() where plan_id = $1 and course_code = $2 returning completed_at`,
			planID, courseCode).Scan(&done); err != nil {
			return err
		}
		return insertAudit(ctx, tx, auditEvent{Actor: actor, Action: "course.completed", EntityType: "learning_plan", EntityID: planID,
			Details: map[string]any{"course_code": courseCode}})
	})
	return done, err
}

// PlanProgress counts the courses in the candidate's latest plan and how many are done.
// A candidate with no plan gets zero and zero.
func (s *Store) PlanProgress(ctx context.Context, candidateID string) (total, completed int, err error) {
	err = s.pool.QueryRow(ctx,
		`select count(m.*), count(m.completed_at)
		 from learning_plan_modules m
		 where m.plan_id = (select id from learning_plans where candidate_id = $1 order by created_at desc limit 1)`,
		candidateID).Scan(&total, &completed)
	return total, completed, err
}
