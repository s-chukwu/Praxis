package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

type fakeCourses struct {
	total, completed int
	lastCode         string
}

func (f *fakeCourses) CompleteCourse(ctx context.Context, candidateID, code string, actor store.Actor) (time.Time, error) {
	f.lastCode = code
	if code == "M-LATER" {
		return time.Time{}, store.ErrWrongState
	}
	return time.Now(), nil
}

func (f *fakeCourses) PlanProgress(ctx context.Context, candidateID string) (int, int, error) {
	return f.total, f.completed, nil
}

const twinCatalogJSON = `{"current":{"scenarios":"1.2.0"},"scenario_banks":[{"version":"1.2.0","scenarios":[
	{"id":"CS-01-A","stage":"baseline","parallel_form_id":"CS-01-B","timebox_minutes":12,"expected_artifacts":["customer_reply"],"candidate_brief":"Baseline brief."},
	{"id":"CS-01-B","stage":"final","parallel_form_id":"CS-01-A","rubric_version":"0.2.0","timebox_minutes":12,"expected_artifacts":["customer_reply"],"candidate_brief":"Final brief."}]}]}`

// newFinalRouter gives candidate A a reviewed CS-01-A baseline on scenarios 1.2.0.
func newFinalRouter(courses *fakeCourses) (http.Handler, *fakeAttempts) {
	attempts := newFakeAttempts()
	a, _ := attempts.CreateAttempt(context.Background(), store.NewAttempt{CandidateID: candidateA, TargetTrack: "operator",
		RoleStandardVersion: "1.0.0", ScenarioVersion: "1.2.0", RubricVersion: "0.2.0", PolicyVersion: "1.0.0", ScenarioID: "CS-01-A"}, store.Actor{})
	a.Status = workflow.Reviewed
	attempts.items[a.ID] = a
	mux := http.NewServeMux()
	registerCourseRoutes(mux, fakeAuth{}, attempts, courses, &fakeCatalog{response: twinCatalogJSON})
	return mux, attempts
}

func TestCoursesCompleteInOrder(t *testing.T) {
	courses := &fakeCourses{}
	h, _ := newFinalRouter(courses)
	expectStatus(t, call(t, h, http.MethodPost, "/v1/me/plan/courses/M-ROLE/complete", candidateA, "candidate", ""), http.StatusOK, "complete")
	if courses.lastCode != "M-ROLE" {
		t.Fatalf("wrong course completed: %q", courses.lastCode)
	}
	expectStatus(t, call(t, h, http.MethodPost, "/v1/me/plan/courses/M-LATER/complete", candidateA, "candidate", ""), http.StatusConflict, "out of order")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/me/plan/courses/bad/complete", candidateA, "candidate", ""), http.StatusNotFound, "bad code")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/me/plan/courses/M-ROLE/complete", reviewerID, "reviewer", ""), http.StatusForbidden, "reviewer")
}

func TestFinalTaskStaysLockedUntilThePlanIsDone(t *testing.T) {
	h, attempts := newFinalRouter(&fakeCourses{total: 3, completed: 2})
	rec := call(t, h, http.MethodGet, "/v1/final/scenario", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusConflict, "locked scenario")
	if strings.Contains(rec.Body.String(), "Final brief") {
		t.Fatalf("the final brief must not show before the plan is done: %s", rec.Body.String())
	}
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/final", candidateA, "candidate", ""), http.StatusConflict, "locked start")
	if len(attempts.items) != 1 {
		t.Fatalf("no final attempt may be created while locked")
	}
}

func TestFinalTaskUsesTheBaselineTwinAndVersions(t *testing.T) {
	h, attempts := newFinalRouter(&fakeCourses{total: 3, completed: 3})
	rec := call(t, h, http.MethodGet, "/v1/final/scenario", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusOK, "scenario")
	if !strings.Contains(rec.Body.String(), `"scenario_id":"CS-01-B"`) || strings.Contains(rec.Body.String(), "Baseline brief") {
		t.Fatalf("expected the CS-01-B twin only, got %s", rec.Body.String())
	}

	rec = call(t, h, http.MethodPost, "/v1/attempts/final", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusCreated, "start")
	for _, want := range []string{`"scenario_id":"CS-01-B"`, `"scenario_version":"1.2.0"`, `"status":"in_progress"`, `"target_track":"operator"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("missing %s in %s", want, rec.Body.String())
		}
	}
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/final", candidateA, "candidate", ""), http.StatusOK, "start again")
	if len(attempts.items) != 2 {
		t.Fatalf("expected one baseline and one final, got %d attempts", len(attempts.items))
	}
}
