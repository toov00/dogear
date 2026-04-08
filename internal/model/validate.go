package model

import (
	"fmt"
	"strings"
)

func ValidPositionType(s string) bool {
	switch strings.TrimSpace(s) {
	case string(PosPage), string(PosChapter), string(PosSection), string(PosLoc), string(PosPercent), string(PosNote):
		return true
	default:
		return false
	}
}

func (d *ExportDoc) Validate() error {
	if d == nil {
		return fmt.Errorf("document is nil")
	}
	v := d.Version
	if v == 0 {
		v = 1
	}
	if v != 1 {
		return fmt.Errorf("unsupported export version %d (expected 1)", d.Version)
	}
	for i := range d.Titles {
		t := &d.Titles[i]
		if strings.TrimSpace(t.Title) == "" {
			return fmt.Errorf("row %d: empty title", i)
		}
		if t.Status != string(StatusActive) && t.Status != string(StatusFinished) {
			return fmt.Errorf("%q: invalid status %q", t.Title, t.Status)
		}
		for j := range t.Checkpoints {
			c := &t.Checkpoints[j]
			if !ValidPositionType(c.PositionType) {
				return fmt.Errorf("%q: invalid position_type %q", t.Title, c.PositionType)
			}
		}
	}
	return nil
}
