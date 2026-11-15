package auth

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// SoftDeleteUser closes a Supabase account so it can never sign in again. A soft delete keeps
// the row, so attempts and scores that reference it stay for the pilot evaluation (PM decision X-04).
func (p *Profiles) SoftDeleteUser(ctx context.Context, userID string) error {
	if !uuidPattern.MatchString(userID) {
		return ErrNoProfile
	}
	endpoint := p.baseURL + "/auth/v1/admin/users/" + url.PathEscape(userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, bytes.NewReader([]byte(`{"should_soft_delete":true}`)))
	if err != nil {
		return err
	}
	req.Header.Set("apikey", p.key)
	if strings.HasPrefix(p.key, "eyJ") {
		req.Header.Set("Authorization", "Bearer "+p.key)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete user: status %d", resp.StatusCode)
	}
	return nil
}
