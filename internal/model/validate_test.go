package model

import (
	"testing"
	"time"
)

func TestExportDocValidate(t *testing.T) {
	d := ExportDoc{Version: 1, Titles: []ExportTitle{{
		ID: 1, Title: "OK", Status: "active", Checkpoints: []ExportCheckpoint{{
			ID: 1, PositionType: "page", CreatedAt: time.Now(),
		}},
	}}}
	if err := d.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestExportDocRejectBadStatus(t *testing.T) {
	d := ExportDoc{Version: 1, Titles: []ExportTitle{{
		ID: 1, Title: "X", Status: "archived",
	}}}
	if err := d.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestExportDocRejectBadCheckpoint(t *testing.T) {
	d := ExportDoc{Version: 1, Titles: []ExportTitle{{
		ID: 1, Title: "X", Status: "active", Checkpoints: []ExportCheckpoint{{
			ID: 1, PositionType: "invalid", CreatedAt: time.Now(),
		}},
	}}}
	if err := d.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
