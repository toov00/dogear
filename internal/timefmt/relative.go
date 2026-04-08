package timefmt

import (
	"fmt"
	"strings"
	"time"
)

func Relative(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	now := time.Now()
	loc := now.Location()
	t = t.In(loc)
	now = now.In(loc)
	if t.After(now) {
		d := t.Sub(now)
		if d < time.Minute {
			return "just now"
		}
		if d < time.Hour {
			m := int(d / time.Minute)
			if m == 1 {
				return "in 1m"
			}
			return fmt.Sprintf("in %dm", m)
		}
		if d < 24*time.Hour {
			h := int(d / time.Hour)
			if h == 1 {
				return "in 1h"
			}
			return fmt.Sprintf("in %dh", h)
		}
		startToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		startT := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		days := int(startT.Sub(startToday).Hours() / 24)
		if days == 1 {
			return "tomorrow"
		}
		if days > 1 && days < 7 {
			return fmt.Sprintf("in %dd", days)
		}
		return t.Format("Jan 2")
	}
	d := now.Sub(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		m := int(d / time.Minute)
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	}
	if d < 24*time.Hour {
		h := int(d / time.Hour)
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	}
	startToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	startPast := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	days := int(startToday.Sub(startPast).Hours() / 24)
	if days == 0 {
		return "today"
	}
	if days == 1 {
		return "yesterday"
	}
	if days < 7 {
		return fmt.Sprintf("%d days ago", days)
	}
	if days < 14 {
		return fmt.Sprintf("%d days ago", days)
	}
	if days < 56 {
		w := days / 7
		if w < 2 {
			return fmt.Sprintf("%d days ago", days)
		}
		return fmt.Sprintf("%d weeks ago", w)
	}
	if days < 365 {
		mo := days / 30
		if mo < 1 {
			mo = 1
		}
		if mo == 1 {
			return "1 month ago"
		}
		if mo >= 12 {
			return t.Format("Jan 2006")
		}
		return fmt.Sprintf("%d months ago", mo)
	}
	y := now.Year() - t.Year()
	if y == 1 || (y == 0 && days >= 365) {
		return "1 year ago"
	}
	if y < 4 && days < 365*4 {
		if y <= 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", y)
	}
	return t.Format("Jan 2006")
}

func TitleLineTime(t time.Time) string {
	s := Relative(t)
	if strings.HasSuffix(s, " ago") || strings.HasPrefix(s, "in ") {
		return "updated " + s
	}
	if s == "today" || s == "yesterday" || s == "tomorrow" {
		return "updated " + s
	}
	if s == "just now" {
		return "updated just now"
	}
	return "updated " + s
}
