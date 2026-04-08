package timefmt

import (
	"testing"
	"time"
)

func TestRelativeToday(t *testing.T) {
	now := time.Now()
	s := Relative(now)
	if s != "today" && s != "just now" {
		t.Fatalf("unexpected %q", s)
	}
}
