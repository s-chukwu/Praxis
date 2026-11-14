package store

import "context"

// AdminUser is one row of the admin user list.
type AdminUser struct {
	ID             string
	FullName       string
	Role           string
	BaselineStatus *string // nil when the user has no baseline attempt
}

// ListUsers returns users with their role and baseline attempt status.
// Admin only (the route checks this). Capped so one call cannot return everything.
func (s *Store) ListUsers(ctx context.Context) ([]AdminUser, error) {
	rows, err := s.pool.Query(ctx,
		`select p.id::text, coalesce(p.full_name, ''), p.role::text,
		        (select a.status::text from attempts a
		          where a.candidate_id = p.id and a.stage = 'baseline' limit 1)
		 from profiles p
		 order by lower(coalesce(p.full_name, '')), p.id
		 limit 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.FullName, &u.Role, &u.BaselineStatus); err != nil {
			return nil, err

		}
		out = append(out, u)
	}
	return out, rows.Err()
}