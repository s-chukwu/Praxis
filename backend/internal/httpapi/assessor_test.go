package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
	"github.com/s-chukwu/Praxis/backend/internal/store"
)

type fakePackets struct {
	err         error
	calls       int
	lastActor   aiclient.Actor
	lastID      string
	lastVersion string
}

func (f *fakePackets) AssessorPacket(ctx context.Context, actor aiclient.Actor, requestID, scenarioID, version string) (json.RawMessage, error) {
	f.calls++
	f.lastActor, f.lastID, f.lastVersion = actor, scenarioID, version
	if f.err != nil {
		return nil, f.err
	}
	return json.RawMessage(`{"scenario":{"id":"CS-01-A"},"key":{"must_include":["finance confirmation"]}}`), nil
}

func newAssessorTestRouter(packets PacketReader) http.Handler {
	attempts := newFakeAttempts()
	reviews := &fakeReviews{
		attempts: attempts,
		reviews:  map[string]store.Review{},
		scores:   map[string][]store.DimensionScore{},
		findings: map[string][]store.Finding{},
	}
	mux := http.NewServeMux()
	registerAttemptRoutes(mux, fakeAuth{}, attempts)
	registerReviewRoutes(mux, fakeAuth{}, reviews)
	registerAssessorRoute(mux, fakeAuth{}, reviews, packets)
	return mux
}

func TestAssessorPacketIsForTheReviewerHoldingTheReview(t *testing.T) {
	packets := &fakePackets{}
	h := newAssessorTestRouter(packets)
	attemptID := submittedAttempt(t, h)
	path := "/v1/attempts/" + attemptID + "/assessor-packet"

	expectStatus(t, call(t, h, http.MethodGet, path, reviewerID, "reviewer", ""), http.StatusNotFound, "before starting a review")
	expectStatus(t, call(t, h, http.MethodGet, path, candidateA, "candidate", ""), http.StatusForbidden, "the candidate")

	startedReview(t, h, attemptID)
	rec := call(t, h, http.MethodGet, path, reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusOK, "reviewer holding the review")
	if !strings.Contains(rec.Body.String(), "must_include") || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected the packet with no-store, got %s %q", rec.Body.String(), rec.Header().Get("Cache-Control"))
	}
	if packets.lastActor != aiclient.ReviewerActor(reviewerID) || packets.lastID != "CS-01-A" || packets.lastVersion != "1.1.0" {
		t.Fatalf("wrong AI call: %q %q %q", packets.lastActor, packets.lastID, packets.lastVersion)
	}

	other := "44444444-4444-4444-4444-444444444444"
	expectStatus(t, call(t, h, http.MethodGet, path, other, "reviewer", ""), http.StatusNotFound, "another reviewer")
	expectStatus(t, call(t, h, http.MethodGet, path, other, "admin", ""), http.StatusOK, "admin")
	if packets.calls != 2 {
		t.Fatalf("the AI service should only be called for allowed callers, got %d calls", packets.calls)
	}
}

func TestAssessorPacketHidesAIFailures(t *testing.T) {
	h := newAssessorTestRouter(&fakePackets{err: errors.New("key vault timeout")})
	attemptID := submittedAttempt(t, h)
	startedReview(t, h, attemptID)

	rec := call(t, h, http.MethodGet, "/v1/attempts/"+attemptID+"/assessor-packet", reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusBadGateway, "ai down")
	if strings.Contains(rec.Body.String(), "key vault") {
		t.Fatalf("internal error details leaked: %s", rec.Body.String())
	}

	h = newAssessorTestRouter(nil)
	expectStatus(t, call(t, h, http.MethodGet, "/v1/attempts/"+attemptID+"/assessor-packet", reviewerID, "reviewer", ""),
		http.StatusServiceUnavailable, "no AI client")
}
