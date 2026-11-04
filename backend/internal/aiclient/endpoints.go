package aiclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

// ErrActorNotAllowed is returned before any request is sent when the caller's
// role may not use that endpoint. The Python service checks again as a second layer.
var ErrActorNotAllowed = errors.New("actor role not allowed for this call")

// kind returns the part before the colon, for example "reviewer" in "reviewer:<uuid>".
func (a Actor) kind() string {
	kind, _, _ := strings.Cut(string(a), ":")
	return kind
}

func requireActor(actor Actor, allowed ...string) error {
	if slices.Contains(allowed, actor.kind()) {
		return nil
	}
	return ErrActorNotAllowed
}

// Compare calls POST /internal/v1/diagnostics/compare (reviewer or system).
// It compares a reviewed baseline with its reviewed parallel final form.
func (c *Client) Compare(ctx context.Context, actor Actor, requestID string, req any) (json.RawMessage, error) {
	if err := requireActor(actor, "reviewer", "system"); err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodPost, "/internal/v1/diagnostics/compare", actor, requestID, req)
}

// AssessorPacket calls GET /internal/v1/scenarios/{id}/assessor-packet (reviewer or admin).
// The answer holds the grading key, so it must never be forwarded to a candidate.
func (c *Client) AssessorPacket(ctx context.Context, actor Actor, requestID, scenarioID, version string) (json.RawMessage, error) {
	if err := requireActor(actor, "reviewer", "admin"); err != nil {
		return nil, err
	}
	path := "/internal/v1/scenarios/" + url.PathEscape(scenarioID) + "/assessor-packet"
	if version != "" {
		path += "?version=" + url.QueryEscape(version)
	}
	return c.do(ctx, http.MethodGet, path, actor, requestID, nil)
}
