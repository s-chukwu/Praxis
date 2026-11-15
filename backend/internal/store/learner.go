package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

// LearnerProfile is the learner's own editable data: profiles plus learner_profiles (migration 009).
type LearnerProfile struct {
	FullName       string
	Programme      *string
	Availability   *string
	OnboardingDone bool
	CVText         *string
	ClaimedSkills  []string
}

// LearnerPatch changes only the fields that are set. ClearCV deletes the CV text without touching the account.
type LearnerPatch struct {
	FullName       *string
	Programme      *string
	Availability   *string
	OnboardingDone *bool
	CVText         *string
	ClearCV        bool
	ClaimedSkills  *[]string
}

func (s *Store) GetLearnerProfile(ctx context.Context, userID string) (LearnerProfile, error) {
	var p LearnerProfile
	var skills []byte
	err := s.pool.QueryRow(ctx,
		`select coalesce(pr.full_name, ''), lp.programme, lp.availability, coalesce(lp.onboarding_done, false),
		        lp.cv_text, coalesce(lp.claimed_skills, '[]'::jsonb)
		 from profiles pr left join learner_profiles lp on lp.user_id = pr.id
		 where pr.id = $1`, userID).Scan(&p.FullName, &p.Programme, &p.Availability, &p.OnboardingDone, &p.CVText, &skills)
	if errors.Is(err, pgx.ErrNoRows) {
		return LearnerProfile{}, ErrNotFound
	}
	if err != nil {
		return LearnerProfile{}, err
	}
	if err := json.Unmarshal(skills, &p.ClaimedSkills); err != nil {
		p.ClaimedSkills = nil
	}
	return p, nil
}

func (s *Store) UpdateLearnerProfile(ctx context.Context, userID string, patch LearnerPatch) error {
	// Sent as text and cast in SQL, so a nil pointer is a clean NULL that keeps the old value.
	var skills *string
	if patch.ClaimedSkills != nil {
		encoded, err := json.Marshal(*patch.ClaimedSkills)
		if err != nil {
			return err
		}
		text := string(encoded)
		skills = &text
	}
	return s.WithTx(ctx, func(q pgx.Tx) error {
		if patch.FullName != nil {
			if _, err := q.Exec(ctx, `update profiles set full_name = $2 where id = $1`, userID, *patch.FullName); err != nil {
				return err
			}
		}
		_, err := q.Exec(ctx,
			`insert into learner_profiles (user_id) values ($1) on conflict (user_id) do nothing`, userID)
		if err != nil {
			return err
		}
		_, err = q.Exec(ctx,
			`update learner_profiles set
			   programme       = coalesce($2, programme),
			   availability    = coalesce($3, availability),
			   onboarding_done = coalesce($4, onboarding_done),
			   cv_text         = case when $6 then null else coalesce($5, cv_text) end,
			   claimed_skills  = coalesce($7::jsonb, claimed_skills),
			   updated_at      = now()
			 where user_id = $1`,
			userID, patch.Programme, patch.Availability, patch.OnboardingDone, patch.CVText, patch.ClearCV, skills)
		return err
	})
}

// EraseLearner removes personal data (PM decision X-04): the learner profile, CV text and name.
// Attempts and scores stay for the pilot evaluation, linked only to an id nobody can sign in with.
func (s *Store) EraseLearner(ctx context.Context, userID string, actor Actor) error {
	return s.WithTx(ctx, func(q pgx.Tx) error {
		if _, err := q.Exec(ctx, `delete from learner_profiles where user_id = $1`, userID); err != nil {
			return err
		}
		if _, err := q.Exec(ctx, `update profiles set full_name = null where id = $1`, userID); err != nil {
			return err
		}
		return insertAudit(ctx, q, auditEvent{Actor: actor, Action: "account.erased", EntityType: "profile", EntityID: userID})
	})
}
