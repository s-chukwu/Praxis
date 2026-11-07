package store

import (
	"context"
	"encoding/json"

	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

// Actor is who is making a change, for the audit log and the role checks.
// ID is empty when the Go backend acts on its own (the system role).
type Actor struct {
	ID   string
	Role workflow.Role
}

type auditEvent struct {
	Actor      Actor
	Action     string
	EntityType string
	EntityID   string
	Details    map[string]any
}

// insertAudit writes one row to the append only audit_events table, inside the
// same transaction as the change it describes. Never put raw CV text, full
// candidate answers, keys or hidden rubric answers in Details.
func insertAudit(ctx context.Context, q querier, e auditEvent) error {
	details := []byte("{}")
	if len(e.Details) > 0 {
		encoded, err := json.Marshal(e.Details)
		if err != nil {
			return err
		}
		details = encoded
	}

	_, err := q.Exec(ctx,
		`insert into audit_events (actor_id, actor_role, action, entity_type, entity_id, details)
		 values ($1, $2, $3, $4, $5, $6)`,
		nullable(e.Actor.ID), string(e.Actor.Role), e.Action, e.EntityType, nullable(e.EntityID), details)
	return err
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
