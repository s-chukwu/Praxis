package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var ErrNoProfile = errors.New("no profile for user")

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F-]{36}$`)

type Profiles struct {
	baseURL string
	key     string
	client  *http.Client
}

type profileRow struct {
	Role string `json:"role"`
}

func NewProfiles(supabaseURL, serviceKey string) *Profiles {
	return &Profiles{
		baseURL: NormalizeBaseURL(supabaseURL),
		key:     serviceKey,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// Role returns the role stored in the profiles table for a user id.
// It uses the service role key, so it must only run on the server.
func (p *Profiles) Role(ctx context.Context, userID string) (string, error) {
	if !uuidPattern.MatchString(userID) {
		return "", ErrNoProfile
	}

	endpoint := p.baseURL + "/rest/v1/profiles?select=role&id=eq." + url.QueryEscape(userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("apikey", p.key)
	if strings.HasPrefix(p.key, "eyJ") {
		req.Header.Set("Authorization", "Bearer "+p.key)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("profile lookup: status %d", resp.StatusCode)
	}

	var rows []profileRow
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&rows); err != nil {
		return "", err
	}
	if len(rows) == 0 || rows[0].Role == "" {
		return "", ErrNoProfile
	}
	return rows[0].Role, nil
}
