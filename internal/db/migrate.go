package db

import (
	"database/sql"
	"fmt"
)

const schema = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS titles (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	author_or_source TEXT NOT NULL DEFAULT '',
	format TEXT NOT NULL DEFAULT '',
	total_pages INTEGER,
	total_chapters INTEGER,
	status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','finished')),
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	finished_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_titles_status_updated ON titles(status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_titles_title_lower ON titles(lower(title));

CREATE TABLE IF NOT EXISTS title_tags (
	title_id INTEGER NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
	tag TEXT NOT NULL,
	PRIMARY KEY (title_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_title_tags_tag ON title_tags(tag);

CREATE TABLE IF NOT EXISTS checkpoints (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title_id INTEGER NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
	position_type TEXT NOT NULL CHECK (position_type IN ('page','chapter','section','loc','percent','note')),
	page INTEGER,
	chapter INTEGER,
	section TEXT,
	loc INTEGER,
	percent REAL,
	note TEXT,
	created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_checkpoints_title_created ON checkpoints(title_id, created_at DESC);
`

func Migrate(conn *sql.DB) error {
	if _, err := conn.Exec(schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
