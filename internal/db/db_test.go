package db

import (
	"path/filepath"
	"testing"
)

func TestOpenCreatesDatabase(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "nested", "dogear.db")
	conn, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	var v int
	if err := conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='titles'").Scan(&v); err != nil || v != 1 {
		t.Fatalf("titles table missing: %v", err)
	}
}
