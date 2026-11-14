package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ErrSelfRoleChange stops an admin from locking themselves out by changing their own role.
var ErrSelfRoleChange = errors.New("an admin cannot change their own role")

// SetUserRole changes a user's role and records who did it. Only candidate and
// reviewer can be set here; admins are created by hand in the database.
func (s *Store) SetUserRole(ctx context.Context, userID, role string, actor Actor) error {
	if role != "candidate" && role != "reviewer" {
		return ErrInvalidInput
	}
	if userID == actor.ID {
		return ErrSelfRoleChange
	}
	return s.WithTx(ctx, func(tx pgx.Tx) error {
		var current string
		err := tx.QueryRow(ctx, `select role::text from profiles where id = $1 for update`, userID).Scan(&current)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if current == role {
			return nil
		}
		if current == "admin" {
			return ErrSelfRoleChange
		}
		if _, err := tx.Exec(ctx, `update profiles set role = $2::user_role where id = $1`, userID, role); err != nil {
			return err
		}
		return insertAudit(ctx, tx, auditEvent{
			Actor:      actor,
			Action:     "profile.role_changed",
			EntityType: "profile",
			EntityID:   userID,
			Details:    map[string]any{"from": current, "to": role},
		})
	})
}
