package store

import "context"

// AdminUser is one row of the admin user list.
type AdminUser struct {
	ID                string
	Email             string
	FullName          string
	Role              string
	BaselineAttemptID *string // nil when the user has no active baseline attempt
	BaselineStatus    *string // nil when the user has no active baseline attempt
}

// ListUsers returns users with their email, role and active baseline attempt.
// A voided baseline attempt is not shown, because the candidate can start again.
// Admin only (the route checks this). Capped so one call cannot return everything.
func (s *Store) ListUsers(ctx context.Context) ([]AdminUser, error) {
	rows, err := s.pool.Query(ctx,
		`select p.id::text, coalesce(u.email, ''), coalesce(p.full_name, ''), p.role::text,
		        b.id::text, b.status::text
		 from profiles p
		 left join auth.users u on u.id = p.id
		 left join lateral (
		   select a.id, a.status from attempts a
		   where a.candidate_id = p.id and a.stage = 'baseline' and a.status <> 'voided'
		   limit 1
		 ) b on true
		 order by lower(coalesce(u.email, '')), p.id
		 limit 500`)
	if err != nil {
		return nil, err

	}
	defer rows.Close()

	var out []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.Role,
			&u.BaselineAttemptID, &u.BaselineStatus); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}