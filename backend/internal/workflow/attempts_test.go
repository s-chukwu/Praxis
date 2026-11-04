package workflow

import (
	"errors"
	"testing"
)

func TestCheckTransition(t *testing.T) {
	cases := []struct {
		from State
		to   State
		role Role
		want error
	}{
		{Assigned, InProgress, Candidate, nil},
		{InProgress, Submitted, Candidate, nil},
		{InProgress, Submitted, System, nil},
		{Submitted, UnderReview, Reviewer, nil},
		{UnderReview, ReviewPending, Reviewer, nil},
		{ReviewPending, UnderReview, Reviewer, nil},
		{UnderReview, Reviewed, Reviewer, nil},
		{Reviewed, UnderReview, Admin, nil},
		{Submitted, Voided, Admin, nil},

		{Submitted, UnderReview, Candidate, ErrRoleNotAllowed},
		{UnderReview, Reviewed, Candidate, ErrRoleNotAllowed},
		{Reviewed, UnderReview, Reviewer, ErrRoleNotAllowed},
		{InProgress, Voided, Reviewer, ErrRoleNotAllowed},
		{Assigned, InProgress, Admin, ErrRoleNotAllowed},

		{Assigned, Reviewed, Admin, ErrIllegalTransition},
		{Reviewed, Submitted, Admin, ErrIllegalTransition},
		{Voided, InProgress, Admin, ErrIllegalTransition},
		{Assigned, Assigned, Admin, ErrIllegalTransition},
	}

	for _, c := range cases {
		err := CheckTransition(c.from, c.to, c.role)
		if c.want == nil {
			if err != nil {
				t.Errorf("%s to %s as %s: expected allowed, got %v", c.from, c.to, c.role, err)
			}
			continue
		}
		if !errors.Is(err, c.want) {
			t.Errorf("%s to %s as %s: expected %v, got %v", c.from, c.to, c.role, c.want, err)
		}
	}
}

func TestVoidedIsFinal(t *testing.T) {
	states := []State{Assigned, InProgress, Submitted, UnderReview, ReviewPending, Reviewed, Voided}
	roles := []Role{Candidate, Reviewer, Employer, Admin, System}

	for _, to := range states {
		for _, role := range roles {
			if err := CheckTransition(Voided, to, role); err == nil {
				t.Errorf("voided to %s as %s must never be allowed", to, role)
			}
		}
	}
}

func TestCheckReviewer(t *testing.T) {
	id := "11111111-1111-1111-1111-111111111111"
	other := "22222222-2222-2222-2222-222222222222"

	if err := CheckReviewer(id, id); !errors.Is(err, ErrSelfReview) {
		t.Fatalf("expected ErrSelfReview, got %v", err)
	}
	if err := CheckReviewer(other, id); err != nil {
		t.Fatalf("a different reviewer should be allowed, got %v", err)
	}
}
