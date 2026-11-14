package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

const reviewerB = "55555555-5555-5555-5555-555555555555"

// fakeReviews is an in memory stand in for the review part of the database layer.
type fakeReviews struct {
	attempts *fakeAttempts
	reviews  map[string]store.Review
	scores   map[string][]store.DimensionScore
	findings map[string][]store.Finding
	count    int
}

func (f *fakeReviews) GetAttempt(ctx context.Context, id string) (store.Attempt, error) {
	return f.attempts.GetAttempt(ctx, id)
}

func (f *fakeReviews) StartReview(ctx context.Context, attemptID string, reviewer store.Actor) (store.Review, error) {
	a, err := f.attempts.GetAttempt(ctx, attemptID)
	if err != nil {
		return store.Review{}, err
	}
	if a.CandidateID == reviewer.ID {
		return store.Review{}, workflow.ErrSelfReview
	}
	if a.Status == workflow.UnderReview {
		for _, r := range f.reviews {
			if r.AttemptID == attemptID && r.FinalizedAt == nil {
				if r.ReviewerID != reviewer.ID {
					return store.Review{}, store.ErrReviewTaken
				}
				return r, nil
			}
		}
	}
	if a.Status != workflow.Submitted && a.Status != workflow.UnderReview {
		return store.Review{}, fmt.Errorf("%w: attempt is %s", store.ErrWrongState, a.Status)
	}

	a.Status = workflow.UnderReview
	f.attempts.items[attemptID] = a

	f.count++
	review := store.Review{
		ID:         fmt.Sprintf("99999999-9999-9999-9999-%012d", f.count),
		AttemptID:  attemptID,
		ReviewerID: reviewer.ID,
		CreatedAt:  time.Now(),
	}
	f.reviews[review.ID] = review
	return review, nil
}

func (f *fakeReviews) ListReviewerReviews(ctx context.Context, reviewerID string) ([]store.Review, error) {
	var out []store.Review
	for _, r := range f.reviews {
		if r.ReviewerID == reviewerID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeReviews) GetReview(ctx context.Context, id string) (store.Review, error) {
	r, ok := f.reviews[id]
	if !ok {
		return store.Review{}, store.ErrNotFound
	}
	return r, nil
}

func (f *fakeReviews) ownDraft(reviewID string, reviewer store.Actor) error {
	r, ok := f.reviews[reviewID]
	if !ok || r.ReviewerID != reviewer.ID {
		return store.ErrReviewNotYours
	}
	if r.FinalizedAt != nil {
		return store.ErrReviewFinalized
	}
	return nil
}

func (f *fakeReviews) SaveDimensionScores(ctx context.Context, reviewID string, reviewer store.Actor, scores []store.ScoreInput) error {
	if err := f.ownDraft(reviewID, reviewer); err != nil {
		return err
	}
	var saved []store.DimensionScore
	for _, s := range scores {
		if s.Score < 0 || s.Score > 4 {
			return fmt.Errorf("%w: score for %s must be 0 to 4", store.ErrInvalidInput, s.Dimension)
		}
		saved = append(saved, store.DimensionScore{Dimension: s.Dimension, Score: s.Score, Evidence: s.Evidence})
	}
	f.scores[reviewID] = saved
	return nil
}

func (f *fakeReviews) SaveCriticalFindings(ctx context.Context, reviewID string, reviewer store.Actor, findings []store.FindingInput) error {
	if err := f.ownDraft(reviewID, reviewer); err != nil {
		return err
	}
	var saved []store.Finding
	for _, x := range findings {
		saved = append(saved, store.Finding{
			EventType:       x.EventType,
			SupportingQuote: x.SupportingQuote,
			PolicyReference: x.PolicyReference,
			Decision:        x.Decision,
		})
	}
	f.findings[reviewID] = saved
	return nil
}

func (f *fakeReviews) ListDimensionScores(ctx context.Context, reviewID string) ([]store.DimensionScore, error) {
	return f.scores[reviewID], nil
}

func (f *fakeReviews) ListCriticalFindings(ctx context.Context, reviewID string) ([]store.Finding, error) {
	return f.findings[reviewID], nil
}

func newReviewTestRouter() (http.Handler, *fakeReviews) {
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
	return mux, reviews
}

// submittedAttempt creates an attempt for candidate A and submits it.
func submittedAttempt(t *testing.T, h http.Handler) string {
	t.Helper()

	id := mustCreate(t, h, candidateA)
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/start", candidateA, "candidate", ""), http.StatusOK, "start")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/submit", candidateA, "candidate", ""), http.StatusOK, "submit")
	return id
}

func startedReview(t *testing.T, h http.Handler, attemptID string) string {
	t.Helper()

	rec := call(t, h, http.MethodPost, "/v1/attempts/"+attemptID+"/review", reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusCreated, "start review")
	var view reviewView
	if err := json.Unmarshal(rec.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	return view.ID
}

const scoresJSON = `{"scores":[{"dimension":"accuracy","score":3,"evidence":["I can help you reset that password now."]}]}`

func TestOnlyReviewersStartReviews(t *testing.T) {
	h, _ := newReviewTestRouter()
	id := submittedAttempt(t, h)

	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/review", candidateA, "candidate", ""), http.StatusForbidden, "candidate starts a review")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/review", adminID, "admin", ""), http.StatusForbidden, "admin starts a review")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/review", "", "", ""), http.StatusUnauthorized, "no login")

	startedReview(t, h, id)
}

func TestReviewCannotStartOnAnUnsubmittedAttempt(t *testing.T) {
	h, _ := newReviewTestRouter()
	id := mustCreate(t, h, candidateA)

	rec := call(t, h, http.MethodPost, "/v1/attempts/"+id+"/review", reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusConflict, "review before submission")
}

func TestReviewerCannotReviewTheirOwnAttempt(t *testing.T) {
	h, _ := newReviewTestRouter()

	// The reviewer is also the candidate of this attempt.
	id := mustCreate(t, h, reviewerID)
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/start", reviewerID, "candidate", ""), http.StatusOK, "start")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/submit", reviewerID, "candidate", ""), http.StatusOK, "submit")

	rec := call(t, h, http.MethodPost, "/v1/attempts/"+id+"/review", reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusForbidden, "self review")
}

func TestOnlyTheOwningReviewerEditsADraft(t *testing.T) {
	h, _ := newReviewTestRouter()
	reviewID := startedReview(t, h, submittedAttempt(t, h))
	path := "/v1/reviews/" + reviewID + "/scores"

	expectStatus(t, call(t, h, http.MethodPut, path, reviewerID, "reviewer", scoresJSON), http.StatusOK, "owner saves scores")
	expectStatus(t, call(t, h, http.MethodPut, path, reviewerB, "reviewer", scoresJSON), http.StatusNotFound, "another reviewer saves scores")
	expectStatus(t, call(t, h, http.MethodPut, path, candidateA, "candidate", scoresJSON), http.StatusForbidden, "candidate saves scores")
	expectStatus(t, call(t, h, http.MethodPut, path, adminID, "admin", scoresJSON), http.StatusForbidden, "admin saves scores")

	bad := `{"scores":[{"dimension":"accuracy","score":9,"evidence":["I can help you reset that password now."]}]}`
	expectStatus(t, call(t, h, http.MethodPut, path, reviewerID, "reviewer", bad), http.StatusUnprocessableEntity, "score out of range")

	expectStatus(t, call(t, h, http.MethodPut, path, reviewerID, "reviewer", `{"surprise":1}`), http.StatusBadRequest, "unknown fields")

	findings := `{"findings":[{"event_type":"unauthorized_action","supporting_quote":"I can help you reset that password now.","policy_reference":"POL-1","decision":"pending"}]}`
	rec := call(t, h, http.MethodPut, "/v1/reviews/"+reviewID+"/findings", reviewerID, "reviewer", findings)
	expectStatus(t, rec, http.StatusOK, "owner saves findings")
	if !strings.Contains(rec.Body.String(), `"decision":"pending"`) || !strings.Contains(rec.Body.String(), `"dimension":"accuracy"`) {
		t.Fatalf("the response should hold the saved scores and findings, got %s", rec.Body.String())
	}
}

func TestFinalizedReviewsAreLocked(t *testing.T) {
	h, reviews := newReviewTestRouter()
	reviewID := startedReview(t, h, submittedAttempt(t, h))

	finalized := time.Now()
	r := reviews.reviews[reviewID]
	r.FinalizedAt = &finalized
	reviews.reviews[reviewID] = r

	rec := call(t, h, http.MethodPut, "/v1/reviews/"+reviewID+"/scores", reviewerID, "reviewer", scoresJSON)
	expectStatus(t, rec, http.StatusConflict, "scores after finalizing")
}

func TestWhoCanReadAReview(t *testing.T) {
	h, reviews := newReviewTestRouter()
	reviewID := startedReview(t, h, submittedAttempt(t, h))
	path := "/v1/reviews/" + reviewID

	expectStatus(t, call(t, h, http.MethodGet, path, reviewerID, "reviewer", ""), http.StatusOK, "owning reviewer")
	expectStatus(t, call(t, h, http.MethodGet, path, adminID, "admin", ""), http.StatusOK, "admin")
	expectStatus(t, call(t, h, http.MethodGet, path, reviewerB, "reviewer", ""), http.StatusNotFound, "another reviewer")
	expectStatus(t, call(t, h, http.MethodGet, path, candidateA, "candidate", ""), http.StatusNotFound, "candidate before it is finalized")
	expectStatus(t, call(t, h, http.MethodGet, path, candidateA, "employer", ""), http.StatusNotFound, "employer")
	expectStatus(t, call(t, h, http.MethodGet, "/v1/reviews/not-a-uuid", reviewerID, "reviewer", ""), http.StatusNotFound, "bad id")

	finalized := time.Now()
	r := reviews.reviews[reviewID]
	r.FinalizedAt = &finalized
	reviews.reviews[reviewID] = r

	expectStatus(t, call(t, h, http.MethodGet, path, candidateA, "candidate", ""), http.StatusOK, "owner candidate after finalizing")
	expectStatus(t, call(t, h, http.MethodGet, path, candidateB, "candidate", ""), http.StatusNotFound, "another candidate after finalizing")
}

func TestReviewRoutesAnswer503WithoutADatabase(t *testing.T) {
	mux := http.NewServeMux()
	registerReviewRoutes(mux, fakeAuth{}, nil)

	rec := call(t, mux, http.MethodGet, "/v1/reviews/"+reviewerID, reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusServiceUnavailable, "no database")
}

func TestStartingAReviewAgainReturnsTheSameDraft(t *testing.T) {
	h, _ := newReviewTestRouter()
	attemptID := submittedAttempt(t, h)
	first := startedReview(t, h, attemptID)
	again := startedReview(t, h, attemptID)
	if again != first {
		t.Fatalf("a reload must return the same draft, got %s then %s", first, again)
	}

	rec := call(t, h, http.MethodPost, "/v1/attempts/"+attemptID+"/review", "44444444-4444-4444-4444-444444444444", "reviewer", "")
	expectStatus(t, rec, http.StatusConflict, "second reviewer")
	if !strings.Contains(rec.Body.String(), "review_taken") {
		t.Fatalf("expected review_taken, got %s", rec.Body.String())
	}
}

func TestReviewersListOnlyTheirOwnReviews(t *testing.T) {
	h, _ := newReviewTestRouter()
	reviewID := startedReview(t, h, submittedAttempt(t, h))

	rec := call(t, h, http.MethodGet, "/v1/reviews", reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusOK, "own list")
	if !strings.Contains(rec.Body.String(), reviewID) {
		t.Fatalf("the reviewer's draft should be listed, got %s", rec.Body.String())
	}

	rec = call(t, h, http.MethodGet, "/v1/reviews", "44444444-4444-4444-4444-444444444444", "reviewer", "")
	expectStatus(t, rec, http.StatusOK, "other reviewer")
	if strings.Contains(rec.Body.String(), reviewID) {
		t.Fatalf("another reviewer's draft leaked: %s", rec.Body.String())
	}
	expectStatus(t, call(t, h, http.MethodGet, "/v1/reviews", candidateA, "candidate", ""), http.StatusForbidden, "candidate")
}
