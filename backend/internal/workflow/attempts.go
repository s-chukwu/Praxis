package workflow

import (
	"errors"
	"fmt"
	"slices"
)

// State is an attempt state. The list matches attempt states version 1.0.0
// published by the AI service, and the attempt_transitions table in the database.
type State string

const (
	Assigned      State = "assigned"
	InProgress    State = "in_progress"
	Submitted     State = "submitted"
	UnderReview   State = "under_review"
	ReviewPending State = "review_pending"
	Reviewed      State = "reviewed"
	Voided        State = "voided"
)

// InitialState is where every new attempt starts.
const InitialState = Assigned

// Role is who is making the move. System is the Go backend acting on its own.
type Role string

const (
	Candidate Role = "candidate"
	Reviewer  Role = "reviewer"
	Employer  Role = "employer"
	Admin     Role = "admin"
	System    Role = "system"
)

var (
	ErrIllegalTransition = errors.New("illegal attempt transition")
	ErrRoleNotAllowed    = errors.New("role not allowed to make this attempt transition")
	ErrSelfReview        = errors.New("a reviewer cannot review their own attempt")
)

type rule struct {
	from  State
	to    State
	roles []Role
}

var rules = []rule{
	{Assigned, InProgress, []Role{Candidate}},
	{InProgress, Submitted, []Role{Candidate, System}},
	{Submitted, UnderReview, []Role{Reviewer}},
	{UnderReview, ReviewPending, []Role{Reviewer}},
	{ReviewPending, UnderReview, []Role{Reviewer}},
	{UnderReview, Reviewed, []Role{Reviewer}},
	{Reviewed, UnderReview, []Role{Admin}},
	{Assigned, Voided, []Role{Admin}},
	{InProgress, Voided, []Role{Admin}},
	{Submitted, Voided, []Role{Admin}},
	{UnderReview, Voided, []Role{Admin}},
	{ReviewPending, Voided, []Role{Admin}},
}

// CheckTransition says whether this role may move an attempt from one state to another.
// Go must call it on every attempt write, in the same transaction as the update.
// The database also blocks illegal moves, but it does not check roles for the service role.
func CheckTransition(from, to State, role Role) error {
	legal := false
	for _, r := range rules {
		if r.from != from || r.to != to {
			continue
		}
		legal = true
		if slices.Contains(r.roles, role) {
			return nil
		}
	}
	if !legal {
		return fmt.Errorf("%w: %s to %s", ErrIllegalTransition, from, to)
	}
	return fmt.Errorf("%w: %s cannot move %s to %s", ErrRoleNotAllowed, role, from, to)
}

// CheckReviewer stops a reviewer from being assigned their own attempt.
// The AI service and the database also reject this.
func CheckReviewer(reviewerID, candidateID string) error {
	if reviewerID == candidateID {
		return ErrSelfReview
	}
	return nil
}
