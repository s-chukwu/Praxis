package workflow

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

type transitionFile struct {
	Source string   `json:"source"`
	Target string   `json:"target"`
	Roles  []string `json:"roles"`
}

type workflowFile struct {
	Version     string           `json:"version"`
	Initial     string           `json:"initial"`
	Transitions []transitionFile `json:"transitions"`
}

// findWorkflowFile looks for attempt_workflow/1.0.0.json anywhere in the repo,
// so the test does not depend on one exact folder layout.
func findWorkflowFile(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}

	var found string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", ".venv", ".next":
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "1.0.0.json" && filepath.Base(filepath.Dir(path)) == "attempt_workflow" {
			found = path
			return filepath.SkipAll
		}
		return nil
	})

	if found == "" {
		t.Fatalf("could not find attempt_workflow/1.0.0.json under %s", root)
	}
	return found
}

func ruleKey(from, to, role string) string {
	return from + "|" + to + "|" + role
}

// TestRulesMatchAttemptWorkflowFile keeps the Go rules from drifting away from
// the published workflow: Go must allow exactly the moves in the file, no more, no less.
func TestRulesMatchAttemptWorkflowFile(t *testing.T) {
	raw, err := os.ReadFile(findWorkflowFile(t))
	if err != nil {
		t.Fatal(err)
	}

	var wf workflowFile
	if err := json.Unmarshal(raw, &wf); err != nil {
		t.Fatalf("read workflow file: %v", err)
	}

	states := []State{Assigned, InProgress, Submitted, UnderReview, ReviewPending, Reviewed, Voided}
	roles := []Role{Candidate, Reviewer, Employer, Admin, System}

	knownStates := map[string]bool{}
	for _, s := range states {
		knownStates[string(s)] = true
	}
	knownRoles := map[string]bool{}
	for _, r := range roles {
		knownRoles[string(r)] = true
	}

	if wf.Initial != string(InitialState) {
		t.Errorf("file starts attempts in %q, Go starts them in %q", wf.Initial, InitialState)
	}

	allowed := map[string]bool{}
	for _, tr := range wf.Transitions {
		if !knownStates[tr.Source] || !knownStates[tr.Target] {
			t.Errorf("file mentions a state Go does not know: %s to %s", tr.Source, tr.Target)
		}
		for _, role := range tr.Roles {
			if !knownRoles[role] {
				t.Errorf("file mentions a role Go does not know: %s", role)
			}
			allowed[ruleKey(tr.Source, tr.Target, role)] = true
		}
	}

	for _, from := range states {
		for _, to := range states {
			for _, role := range roles {
				want := allowed[ruleKey(string(from), string(to), string(role))]
				err := CheckTransition(from, to, role)

				if want && err != nil {
					t.Errorf("file allows %s to %s for %s, but Go refuses: %v", from, to, role, err)
				}
				if !want && err == nil {
					t.Errorf("Go allows %s to %s for %s, but the file does not", from, to, role)
				}
			}
		}
	}
}
