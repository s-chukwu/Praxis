package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// These tests only run when DATABASE_URL is set, so CI and the local stack
// stay green without a database.
func connectForTest(t *testing.T) *Store {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s, err := New(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(s.Close)
	return s
}

func TestConnectsAndReadsRoles(t *testing.T) {
	s := connectForTest(t)

	var count int
	err := s.WithTx(context.Background(), func(tx pgx.Tx) error {
		return tx.QueryRow(context.Background(), "select count(*) from profiles").Scan(&count)
	})
	if err != nil {
		t.Fatalf("query profiles: %v", err)
	}
	t.Logf("profiles rows: %d", count)
}

func TestWithTxRollsBackOnError(t *testing.T) {
	s := connectForTest(t)
	ctx := context.Background()

	// Insert a consent row for an existing profile, then fail on purpose.
	var profileID string
	if err := s.WithTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, "select id from profiles limit 1").Scan(&profileID)
	}); err != nil {
		t.Skipf("no profile to test with: %v", err)
	}

	errBoom := context.Canceled
	err := s.WithTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			"insert into consents (profile_id, type, granted) values ($1, 'data_processing', true)",
			profileID); err != nil {
			return err
		}
		return errBoom
	})
	if err != errBoom {
		t.Fatalf("expected the callback error back, got %v", err)
	}

	var count int
	if err := s.WithTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"select count(*) from consents where profile_id = $1 and recorded_at > now() - interval '1 minute'",
			profileID).Scan(&count)
	}); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("the insert should have been rolled back, found %d rows", count)
	}
}
