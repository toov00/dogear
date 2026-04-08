package db

import (
	"path/filepath"
	"testing"
)

func TestResolvePathUsesFlag(t *testing.T) {
	p, err := ResolvePath("/tmp/explicit.db")
	if err != nil {
		t.Fatal(err)
	}
	if p != "/tmp/explicit.db" {
		t.Fatalf("got %q", p)
	}
}

func TestResolvePathUsesEnv(t *testing.T) {
	t.Setenv("DOGEAR_DB", "/var/tmp/fromenv.db")
	p, err := ResolvePath("")
	if err != nil {
		t.Fatal(err)
	}
	if p != "/var/tmp/fromenv.db" {
		t.Fatalf("got %q", p)
	}
}

func TestResolvePathDefaultLayout(t *testing.T) {
	t.Setenv("DOGEAR_DB", "")
	p, err := ResolvePath("")
	if err != nil || p == "" {
		t.Fatalf("err %v path %q", err, p)
	}
	if filepath.Base(p) != "dogear.db" {
		t.Fatalf("unexpected file name in %q", p)
	}
	base := filepath.Base(filepath.Dir(p))
	if base != "dogear" {
		t.Fatalf("expected …/dogear/dogear.db, got %q", p)
	}
}
