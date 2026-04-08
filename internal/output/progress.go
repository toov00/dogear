package output

import "dogear/internal/model"

func ProgressPercent(cp *model.Checkpoint, title *model.Title) *float64 {
	if cp == nil {
		return nil
	}
	if cp.Percent != nil {
		p := *cp.Percent
		return &p
	}
	var total int
	if cp.TotalPagesRef != nil && *cp.TotalPagesRef > 0 {
		total = *cp.TotalPagesRef
	} else if title != nil && title.TotalPages != nil && *title.TotalPages > 0 {
		total = *title.TotalPages
	}
	if total <= 0 {
		return nil
	}
	if cp.Page != nil && *cp.Page > 0 {
		p := float64(*cp.Page) / float64(total) * 100
		if p > 100 {
			p = 100
		}
		return &p
	}
	return nil
}
