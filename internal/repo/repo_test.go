package repo

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"dogear/internal/db"
	"dogear/internal/model"
	"dogear/internal/position"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	p := filepath.Join(t.TempDir(), "d.db")
	conn, err := db.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return New(conn)
}

func TestAddCheckpointHistory(t *testing.T) {
	rp := newTestStore(t)
	ctx := context.Background()
	now := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	id, err := rp.InsertTitle(ctx, model.TitleInput{Title: "Book", Format: "paperback", TotalPages: intPtr(100)}, now)
	if err != nil {
		t.Fatal(err)
	}
	cp1, err := position.BuildInput(intPtr(10), nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rp.InsertCheckpoint(ctx, id, cp1, now); err != nil {
		t.Fatal(err)
	}
	cp2, err := position.BuildInput(intPtr(25), nil, nil, nil, nil, strPtr("mid"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rp.InsertCheckpoint(ctx, id, cp2, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	h, err := rp.History(ctx, id)
	if err != nil || len(h) != 2 {
		t.Fatalf("history len %d err %v", len(h), err)
	}
	if h[0].Page == nil || *h[0].Page != 25 {
		t.Fatalf("newest first expected page 25, got %+v", h[0])
	}
	full, err := rp.LoadTitleFull(ctx, id)
	if err != nil || full.LatestCheckpoint == nil || full.LatestCheckpoint.Page == nil || *full.LatestCheckpoint.Page != 25 {
		t.Fatalf("latest %+v err %v", full, err)
	}
}

func TestFinishAndDelete(t *testing.T) {
	rp := newTestStore(t)
	ctx := context.Background()
	now := time.Now()
	id, err := rp.InsertTitle(ctx, model.TitleInput{Title: "Paper"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.Finish(ctx, id, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	full, err := rp.LoadTitleFull(ctx, id)
	if err != nil || full.Status != model.StatusFinished || full.FinishedAt == nil {
		t.Fatalf("finish state %+v err %v", full, err)
	}
	if err := rp.DeleteTitle(ctx, id); err != nil {
		t.Fatal(err)
	}
	full, err = rp.LoadTitleFull(ctx, id)
	if err != nil || full != nil {
		t.Fatalf("expected nil title, got %+v", full)
	}
}

func TestDuplicateActive(t *testing.T) {
	rp := newTestStore(t)
	ctx := context.Background()
	now := time.Now()
	if _, err := rp.InsertTitle(ctx, model.TitleInput{Title: "X", Format: "pdf"}, now); err != nil {
		t.Fatal(err)
	}
	ok, err := rp.HasDuplicateActive(ctx, "x", "pdf")
	if err != nil || !ok {
		t.Fatalf("dup %v err %v", ok, err)
	}
}

func TestImportExportRoundtrip(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.db")
	c1, err := db.Open(p1)
	if err != nil {
		t.Fatal(err)
	}
	s1 := New(c1)
	ctx := context.Background()
	now := time.Now()
	id, err := s1.InsertTitle(ctx, model.TitleInput{Title: "Echo", Format: "epub", TotalPages: intPtr(200)}, now)
	if err != nil {
		t.Fatal(err)
	}
	cp, err := position.BuildInput(intPtr(5), nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s1.InsertCheckpoint(ctx, id, cp, now); err != nil {
		t.Fatal(err)
	}
	if err := s1.AddTag(ctx, id, "fiction"); err != nil {
		t.Fatal(err)
	}
	doc, err := s1.ExportSnapshot(ctx)
	_ = c1.Close()
	if err != nil || len(doc.Titles) != 1 {
		t.Fatalf("export %v err %v", doc, err)
	}
	p2 := filepath.Join(dir, "b.db")
	c2, err := db.Open(p2)
	if err != nil {
		t.Fatal(err)
	}
	s2 := New(c2)
	defer c2.Close()
	if err := s2.ImportReplace(ctx, doc); err != nil {
		t.Fatal(err)
	}
	all, err := s2.LoadAllTitles(ctx)
	if err != nil || len(all) != 1 || all[0].Title != "Echo" {
		t.Fatalf("after import %+v err %v", all, err)
	}
	h, err := s2.History(ctx, all[0].ID)
	if err != nil || len(h) != 1 {
		t.Fatalf("checkpoints %d err %v", len(h), err)
	}
}

func TestImportRejectInvalidDoc(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "z.db")
	c, err := db.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	st := New(c)
	ctx := context.Background()
	doc := &model.ExportDoc{Version: 1, Titles: []model.ExportTitle{{ID: 1, Title: "x", Status: "archived"}}}
	if err := st.ImportReplace(ctx, doc); err == nil {
		t.Fatal("expected validation error")
	}
}

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }
