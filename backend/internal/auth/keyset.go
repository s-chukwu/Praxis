package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"
)

const minRefreshInterval = 30 * time.Second

var (
	ErrKeysUnavailable = errors.New("signing keys unavailable")
	errUnknownKey      = errors.New("unknown signing key")
)

type jwk struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// KeySet holds the public keys Supabase publishes for checking login tokens.
// Keys are fetched on first use and again when an unknown key id appears
// (for example after a key rotation), at most once every 30 seconds.
type KeySet struct {
	url    string
	client *http.Client
	now    func() time.Time

	mu        sync.Mutex
	keys      map[string]*ecdsa.PublicKey
	fetchedAt time.Time
}

func NewKeySet(url string) *KeySet {
	return &KeySet{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
		now:    time.Now,
		keys:   map[string]*ecdsa.PublicKey{},
	}
}

func (k *KeySet) Key(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if key, ok := k.keys[kid]; ok {
		return key, nil
	}
	if !k.fetchedAt.IsZero() && k.now().Sub(k.fetchedAt) < minRefreshInterval {
		return nil, errUnknownKey
	}
	if err := k.refreshLocked(ctx); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrKeysUnavailable, err)
	}
	if key, ok := k.keys[kid]; ok {
		return key, nil
	}
	return nil, errUnknownKey
}

func (k *KeySet) refreshLocked(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.url, nil)
	if err != nil {
		return err
	}
	resp, err := k.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks fetch: status %d", resp.StatusCode)
	}

	var doc struct {
		Keys []jwk `json:"keys"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&doc); err != nil {
		return err
	}

	keys := make(map[string]*ecdsa.PublicKey)
	for _, j := range doc.Keys {
		if j.Kty != "EC" || j.Crv != "P-256" || j.Kid == "" {
			continue
		}
		xb, err := base64.RawURLEncoding.DecodeString(j.X)
		if err != nil {
			continue
		}
		yb, err := base64.RawURLEncoding.DecodeString(j.Y)
		if err != nil {
			continue
		}
		x := new(big.Int).SetBytes(xb)
		y := new(big.Int).SetBytes(yb)
		if !elliptic.P256().IsOnCurve(x, y) {
			continue
		}
		keys[j.Kid] = &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	}
	if len(keys) == 0 {
		return errors.New("jwks contains no usable keys")
	}

	k.keys = keys
	k.fetchedAt = k.now()
	return nil
}
