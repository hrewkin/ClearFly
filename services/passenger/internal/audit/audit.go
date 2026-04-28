// Package audit records sensitive user actions (logins, refunds, profile
// edits, …) into a Postgres-backed audit_log table. Entries are append-only
// and intended to be consumed by the admin "Аудит" page and any future
// off-line analysis.
package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Entry is a single audit record.
type Entry struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ActorID    uuid.UUID `json:"actor_id" db:"actor_id"`
	ActorRole  string    `json:"actor_role" db:"actor_role"`
	ActorName  string    `json:"actor_name" db:"actor_name"`
	Action     string    `json:"action" db:"action"`
	TargetType string    `json:"target_type" db:"target_type"`
	TargetID   string    `json:"target_id" db:"target_id"`
	Details    string    `json:"details" db:"details"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Repository is the audit log backing store.
type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository { return &Repository{db: db} }

// Migrate creates the audit_log table if missing.
func (r *Repository) Migrate(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS audit_log (
		id UUID PRIMARY KEY,
		actor_id UUID NOT NULL,
		actor_role VARCHAR(20) NOT NULL,
		actor_name VARCHAR(255) NOT NULL DEFAULT '',
		action VARCHAR(64) NOT NULL,
		target_type VARCHAR(40) NOT NULL DEFAULT '',
		target_id VARCHAR(64) NOT NULL DEFAULT '',
		details TEXT NOT NULL DEFAULT '',
		ip_address VARCHAR(64) NOT NULL DEFAULT '',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS audit_log_created_at_idx ON audit_log (created_at DESC);
	CREATE INDEX IF NOT EXISTS audit_log_actor_idx ON audit_log (actor_id);`
	_, err := r.db.ExecContext(ctx, schema)
	return err
}

// Log inserts a new audit entry. `details` may be nil.
func (r *Repository) Log(ctx context.Context, e *Entry) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	q := `INSERT INTO audit_log (id, actor_id, actor_role, actor_name, action, target_type, target_id, details, ip_address, created_at)
	      VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.ExecContext(ctx, q,
		e.ID, e.ActorID, e.ActorRole, e.ActorName, e.Action,
		e.TargetType, e.TargetID, e.Details, e.IPAddress, e.CreatedAt)
	return err
}

// LogJSON is a small helper that serializes the supplied detail object as
// JSON before storing it. Logging errors are swallowed because audit must
// never block the request path.
func (r *Repository) LogJSON(ctx context.Context, e Entry, details interface{}) {
	if details != nil {
		if b, err := json.Marshal(details); err == nil {
			e.Details = string(b)
		}
	}
	_ = r.Log(ctx, &e)
}

// List returns the latest entries (newest first), capped at limit (≤500).
func (r *Repository) List(ctx context.Context, limit int) ([]Entry, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	var out []Entry
	q := `SELECT id, actor_id, actor_role, actor_name, action, target_type, target_id, details, ip_address, created_at
	      FROM audit_log ORDER BY created_at DESC LIMIT $1`
	if err := r.db.SelectContext(ctx, &out, q, limit); err != nil {
		return nil, err
	}
	return out, nil
}
