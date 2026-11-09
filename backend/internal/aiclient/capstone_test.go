package aiclient

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
)

func TestCapstoneEvaluateCallerRules(t *testing.T) {
	var calls int32
	srv := newSigningServer(t, &calls, "/internal/v1/capstones/evaluate")

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	const candidateID = "11111111-1111-1111-1111-111111111111"
	const otherID = "22222222-2222-2222-2222-222222222222"

	request := func(suite, candidate string) CapstoneRequest {
		return CapstoneRequest{
			RequestID:   "r1",
			AttemptID:   "33333333-3333-3333-3333-333333333333",
			CandidateID: candidate,
			Suite:       suite,
			Config:      json.RawMessage(`{}`),
		}
	}

	refused := []struct {
		name  string
		actor Actor
		req   CapstoneRequest
		want  error
	}{
		{"candidate cannot run held out", CandidateActor(candidateID), request(SuiteHeldOut, candidateID), ErrActorNotAllowed},
		{"candidate cannot run for someone else", CandidateActor(otherID), request(SuitePractice, candidateID), ErrActorNotAllowed},
		{"admin is not allowed", AdminActor(candidateID), request(SuitePractice, candidateID), ErrActorNotAllowed},
		{"unknown suite is refused", SystemActor, request("everything", candidateID), ErrInvalidSuite},
	}
	for _, c := range refused {
		if _, err := client.CapstoneEvaluate(ctx, c.actor, "", c.req); !errors.Is(err, c.want) {
			t.Errorf("%s: expected %v, got %v", c.name, c.want, err)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("refused calls must not reach the service, got %d calls", got)
	}

	allowed := []struct {
		name  string
		actor Actor
		req   CapstoneRequest
	}{
		{"candidate runs own practice", CandidateActor(candidateID), request(SuitePractice, candidateID)},
		{"reviewer runs held out", ReviewerActor(otherID), request(SuiteHeldOut, candidateID)},
		{"system runs held out", SystemActor, request(SuiteHeldOut, candidateID)},
	}
	for _, c := range allowed {
		if _, err := client.CapstoneEvaluate(ctx, c.actor, "", c.req); err != nil {
			t.Errorf("%s: should be allowed, got %v", c.name, err)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 calls, got %d", got)
	}
}
