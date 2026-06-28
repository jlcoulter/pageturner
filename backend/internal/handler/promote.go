package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PromoteHandler triggers the staging-to-production promotion via import.sql logic
type PromoteHandler struct {
	pool *pgxpool.Pool
}

func NewPromoteHandler(pool *pgxpool.Pool) *PromoteHandler {
	return &PromoteHandler{pool: pool}
}

func (h *PromoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slog.Info("starting staging-to-production promote")

	ctx := r.Context()

	if err := h.promote(ctx); err != nil {
		slog.Error("promote failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	slog.Info("promote complete")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (h *PromoteHandler) promote(ctx context.Context) error {
	// Step 0: Ensure pg_trgm is available
	if _, err := h.pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pg_trgm`); err != nil {
		return err
	}

	// Step 1: Session tuning
	sessionSettings := []string{
		`SET maintenance_work_mem = '1GB'`,
		`SET work_mem = '256MB'`,
		`SET synchronous_commit = OFF`,
	}
	for _, s := range sessionSettings {
		if _, err := h.pool.Exec(ctx, s); err != nil {
			slog.Warn("session setting failed", "sql", s, "error", err)
		}
	}

	// Step 2: Drop previous _new tables if they exist
	dropTables := []string{
		`DROP TABLE IF EXISTS openlibrary.search_documents_new`,
		`DROP TABLE IF EXISTS openlibrary.work_authors_new`,
		`DROP TABLE IF EXISTS openlibrary.authors_new`,
		`DROP TABLE IF EXISTS openlibrary.works_new`,
	}
	for _, d := range dropTables {
		if _, err := h.pool.Exec(ctx, d); err != nil {
			return err
		}
	}

	// Step 3: Create UNLOGGED build tables
	createTables := []string{
		`CREATE UNLOGGED TABLE openlibrary.authors_new (
			id TEXT PRIMARY KEY,
			author_name TEXT NOT NULL,
			search_vector tsvector GENERATED ALWAYS AS (
				to_tsvector('english', author_name)
			) STORED
		)`,
		`CREATE UNLOGGED TABLE openlibrary.works_new (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			search_vector tsvector GENERATED ALWAYS AS (
				to_tsvector('english', title)
			) STORED
		)`,
		`CREATE UNLOGGED TABLE openlibrary.work_authors_new (
			work_id TEXT NOT NULL,
			author_id TEXT NOT NULL,
			PRIMARY KEY (work_id, author_id)
		)`,
	}
	for _, c := range createTables {
		if _, err := h.pool.Exec(ctx, c); err != nil {
			return err
		}
	}

	// Step 4: Bulk load authors from staging
	slog.Info("promote: loading authors from staging")
	start := time.Now()
	tag, err := h.pool.Exec(ctx, `
		INSERT INTO openlibrary.authors_new (id, author_name)
		SELECT id, author_name
		FROM openlibrary.authors_stage
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return err
	}
	slog.Info("promote: authors loaded", "rows", tag.RowsAffected(), "duration", time.Since(start))

	// Step 5: Bulk load works from staging
	slog.Info("promote: loading works from staging")
	start = time.Now()
	tag, err = h.pool.Exec(ctx, `
		INSERT INTO openlibrary.works_new (id, title)
		SELECT id, title
		FROM openlibrary.works_stage
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return err
	}
	slog.Info("promote: works loaded", "rows", tag.RowsAffected(), "duration", time.Since(start))

	// Step 6: Bulk load relationships from staging
	slog.Info("promote: loading work-author relationships from staging")
	start = time.Now()
	tag, err = h.pool.Exec(ctx, `
		INSERT INTO openlibrary.work_authors_new (work_id, author_id)
		SELECT wa.work_id, wa.author_id
		FROM openlibrary.work_authors_stage wa
		JOIN openlibrary.works_new w ON w.id = wa.work_id
		JOIN openlibrary.authors_new a ON a.id = wa.author_id
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return err
	}
	slog.Info("promote: work-authors loaded", "rows", tag.RowsAffected(), "duration", time.Since(start))

	// Step 7: Create search documents
	slog.Info("promote: building search documents")
	start = time.Now()
	if _, err := h.pool.Exec(ctx, `
		CREATE UNLOGGED TABLE openlibrary.search_documents_new (
			work_id TEXT PRIMARY KEY,
			title TEXT,
			author_names TEXT,
			search_vector tsvector
		)
	`); err != nil {
		return err
	}

	tag, err = h.pool.Exec(ctx, `
		INSERT INTO openlibrary.search_documents_new
		SELECT
			w.id,
			w.title,
			string_agg(a.author_name, ' ') AS author_names,
			setweight(to_tsvector('english', w.title), 'A') ||
			setweight(to_tsvector('english', string_agg(a.author_name, ' ')), 'B')
		FROM openlibrary.works_new w
		JOIN openlibrary.work_authors_new wa
			ON w.id = wa.work_id
		JOIN openlibrary.authors_new a
			ON wa.author_id = a.id
		GROUP BY w.id, w.title
	`)
	if err != nil {
		return err
	}
	slog.Info("promote: search documents built", "rows", tag.RowsAffected(), "duration", time.Since(start))

	// Step 8: Convert to logged tables
	slog.Info("promote: converting to logged tables")
	for _, alter := range []string{
		`ALTER TABLE openlibrary.authors_new SET LOGGED`,
		`ALTER TABLE openlibrary.works_new SET LOGGED`,
		`ALTER TABLE openlibrary.work_authors_new SET LOGGED`,
		`ALTER TABLE openlibrary.search_documents_new SET LOGGED`,
	} {
		if _, err := h.pool.Exec(ctx, alter); err != nil {
			return err
		}
	}

	// Step 9: Create indexes
	slog.Info("promote: creating indexes")
	indexes := []string{
		`CREATE INDEX idx_authors_name_new ON openlibrary.authors_new (author_name)`,
		`CREATE INDEX idx_works_title_new ON openlibrary.works_new (title)`,
		`CREATE INDEX idx_work_authors_work_new ON openlibrary.work_authors_new (work_id)`,
		`CREATE INDEX idx_work_authors_author_new ON openlibrary.work_authors_new (author_id)`,
		`CREATE INDEX idx_search_documents_vector_new ON openlibrary.search_documents_new USING GIN(search_vector)`,
		`CREATE INDEX idx_search_documents_title_trgm_new ON openlibrary.search_documents_new USING GIN (title gin_trgm_ops)`,
		`CREATE INDEX idx_search_documents_authors_trgm_new ON openlibrary.search_documents_new USING GIN (author_names gin_trgm_ops)`,
	}
	for _, idx := range indexes {
		start := time.Now()
		if _, err := h.pool.Exec(ctx, idx); err != nil {
			slog.Warn("index creation failed", "sql", idx, "error", err)
		} else {
			slog.Info("promote: index created", "duration", time.Since(start))
		}
	}

	// Step 10: Add foreign keys
	slog.Info("promote: adding foreign keys")
	fks := []string{
		`ALTER TABLE openlibrary.work_authors_new
		ADD CONSTRAINT fk_work
		FOREIGN KEY (work_id)
		REFERENCES openlibrary.works_new(id)
		ON DELETE CASCADE
		NOT VALID`,
		`ALTER TABLE openlibrary.work_authors_new
		ADD CONSTRAINT fk_author
		FOREIGN KEY (author_id)
		REFERENCES openlibrary.authors_new(id)
		ON DELETE CASCADE
		NOT VALID`,
	}
	for _, fk := range fks {
		if _, err := h.pool.Exec(ctx, fk); err != nil {
			slog.Warn("foreign key creation failed", "error", err)
		}
	}

	// Step 11: Validate constraints
	validations := []string{
		`ALTER TABLE openlibrary.work_authors_new VALIDATE CONSTRAINT fk_work`,
		`ALTER TABLE openlibrary.work_authors_new VALIDATE CONSTRAINT fk_author`,
	}
	for _, v := range validations {
		if _, err := h.pool.Exec(ctx, v); err != nil {
			slog.Warn("constraint validation failed", "error", err)
		}
	}

	// Step 12: Atomic table swap
	slog.Info("promote: swapping tables")
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	swaps := []string{
		`ALTER TABLE openlibrary.search_documents RENAME TO search_documents_old`,
		`ALTER TABLE openlibrary.work_authors RENAME TO work_authors_old`,
		`ALTER TABLE openlibrary.authors RENAME TO authors_old`,
		`ALTER TABLE openlibrary.works RENAME TO works_old`,
		`ALTER TABLE openlibrary.authors_new RENAME TO authors`,
		`ALTER TABLE openlibrary.works_new RENAME TO works`,
		`ALTER TABLE openlibrary.work_authors_new RENAME TO work_authors`,
		`ALTER TABLE openlibrary.search_documents_new RENAME TO search_documents`,
	}
	for _, s := range swaps {
		if _, err := tx.Exec(ctx, s); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	slog.Info("promote: tables swapped successfully")

	// Step 13: Cleanup old tables
	slog.Info("promote: cleaning up old tables")
	drops := []string{
		`DROP TABLE IF EXISTS openlibrary.search_documents_old`,
		`DROP TABLE IF EXISTS openlibrary.work_authors_old`,
		`DROP TABLE IF EXISTS openlibrary.authors_old`,
		`DROP TABLE IF EXISTS openlibrary.works_old`,
	}
	for _, d := range drops {
		if _, err := h.pool.Exec(ctx, d); err != nil {
			slog.Warn("drop old table failed", "sql", d, "error", err)
		}
	}

	// Step 14: Reset session settings
	resets := []string{
		`RESET maintenance_work_mem`,
		`RESET work_mem`,
		`RESET synchronous_commit`,
		`RESET temp_buffers`,
	}
	for _, r := range resets {
		if _, err := h.pool.Exec(ctx, r); err != nil {
			slog.Warn("reset setting failed", "sql", r, "error", err)
		}
	}

	slog.Info("promote: complete")
	return nil
}