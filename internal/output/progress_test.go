package output

import (
	"testing"

	"dogear/internal/model"
)

func TestProgressFromPages(t *testing.T) {
	tp := 310
	pg := 61
	ti := &model.Title{TotalPages: &tp}
	cp := &model.Checkpoint{Page: &pg, PositionType: model.PosPage}
	p := ProgressPercent(cp, ti)
	if p == nil || *p < 19.6 || *p > 19.7 {
		t.Fatalf("expected ~19.7%%, got %v", p)
	}
}

func TestProgressExplicitPercent(t *testing.T) {
	pct := 42.0
	cp := &model.Checkpoint{Percent: &pct, PositionType: model.PosPercent}
	p := ProgressPercent(cp, nil)
	if p == nil || *p != 42 {
		t.Fatalf("got %v", p)
	}
}
