package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	minSecretLength  = 32
	maxRetries       = 2
	requestTimeout   = 2 * time.Second
	retryBackoff     = 200 * time.Millisecond
	maxResponseBytes = 4 << 20
)

// Actor is who Go says is behind the call, for example "reviewer:<uuid>".
type Actor string

const SystemActor Actor = "system"

func ReviewerActor(id string) Actor  { return Actor("reviewer:" + id) }
func CandidateActor(id string) Actor { return Actor("candidate:" + id) }
func AdminActor(id string) Actor     { return Actor("admin:" + id) }

// APIError is the error shape returned by the Python service.
type APIError struct {
	Status    int
	Code      string
	Message   string
	RequestID string
	Details   json.RawMessage
}

func (e *APIError) Error() string {
	return fmt.Sprintf("ai service error: status %d, code %q, message %q", e.Status, e.Code, e.Message)
}

type Client struct {
	baseURL string
	secret  string
	http    *http.Client
	now     func() time.Time
}

func New(baseURL, secret string) (*Client, error) {
	if len(secret) < minSecretLength {
		return nil, errors.New("AI_SERVICE_SHARED_SECRET must be at least 32 bytes")
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		secret:  secret,
		http:    &http.Client{Timeout: requestTimeout},
		now:     time.Now,
	}, nil
}

// Catalog calls GET /internal/v1/catalog.
func (c *Client) Catalog(ctx context.Context, actor Actor, requestID string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, "/internal/v1/catalog", actor, requestID, nil)
}

// ScoreDiagnostic calls POST /internal/v1/diagnostics/score (reviewer only).
func (c *Client) ScoreDiagnostic(ctx context.Context, actor Actor, requestID string, req any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, "/internal/v1/diagnostics/score", actor, requestID, req)
}

// LearningPlan calls POST /internal/v1/learning-plans (reviewer or system).
func (c *Client) LearningPlan(ctx context.Context, actor Actor, requestID string, req any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, "/internal/v1/learning-plans", actor, requestID, req)
}

// do retries connection errors and 5xx responses up to maxRetries times, never 4xx.
// Every attempt is signed again so the timestamp stays fresh.
func (c *Client) do(ctx context.Context, method, path string, actor Actor, requestID string, in any) (json.RawMessage, error) {
	var body []byte
	if in != nil {
		encoded, err := json.Marshal(in)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
		body = encoded
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * retryBackoff):
			}
		}

		data, retry, err := c.once(ctx, method, path, actor, requestID, body)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, lastErr
}

func (c *Client) once(ctx context.Context, method, path string, actor Actor, requestID string, body []byte) (json.RawMessage, bool, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, false, err
	}

	timestamp := c.now().Unix()
	req.Header.Set("X-Praxis-Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-Praxis-Actor", string(actor))
	req.Header.Set("X-Praxis-Signature", sign(c.secret, timestamp, method, path, string(actor), body))
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, ctx.Err() == nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, ctx.Err() == nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return data, false, nil
	}
	return nil, resp.StatusCode >= 500, parseAPIError(resp.StatusCode, data)
}

func parseAPIError(status int, data []byte) *APIError {
	var envelope struct {
		Error struct {
			Code      string          `json:"code"`
			Message   string          `json:"message"`
			RequestID string          `json:"request_id"`
			Details   json.RawMessage `json:"details"`
		} `json:"error"`
	}

	apiErr := &APIError{Status: status}
	if err := json.Unmarshal(data, &envelope); err == nil {
		apiErr.Code = envelope.Error.Code
		apiErr.Message = envelope.Error.Message
		apiErr.RequestID = envelope.Error.RequestID
		apiErr.Details = envelope.Error.Details
	}
	return apiErr
}
