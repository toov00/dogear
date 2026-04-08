package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"dogear/internal/model"
	"dogear/internal/search"
	"dogear/internal/timefmt"
)

const listColTitle = 36
const listColPos = 26

func Where(w io.Writer, t *model.Title) {
	if t == nil {
		return
	}
	fmt.Fprintln(w, t.Title)
	line := PositionSummary(t.LatestCheckpoint, t)
	if line != "" {
		fmt.Fprintln(w, line)
	}
	fmt.Fprintln(w, timefmt.TitleLineTime(t.UpdatedAt))
	if t.LatestCheckpoint != nil && t.LatestCheckpoint.Note != nil && strings.TrimSpace(*t.LatestCheckpoint.Note) != "" {
		fmt.Fprintf(w, "note: %s\n", strings.TrimSpace(*t.LatestCheckpoint.Note))
	}
}

func PositionSummary(cp *model.Checkpoint, t *model.Title) string {
	base := checkpointCore(cp, t)
	if base == "" {
		if pct := ProgressPercent(cp, t); pct != nil {
			return fmt.Sprintf("%.0f%%", *pct)
		}
		return ""
	}
	if pct := ProgressPercent(cp, t); pct != nil {
		return base + fmt.Sprintf(" • %.0f%%", *pct)
	}
	return base
}

func checkpointCore(cp *model.Checkpoint, t *model.Title) string {
	if cp == nil {
		return ""
	}
	switch cp.PositionType {
	case model.PosPage:
		if cp.Page != nil {
			if t != nil && t.TotalPages != nil && *t.TotalPages > 0 {
				return fmt.Sprintf("page %d / %d", *cp.Page, *t.TotalPages)
			}
			if cp.TotalPagesRef != nil && *cp.TotalPagesRef > 0 {
				return fmt.Sprintf("page %d / %d", *cp.Page, *cp.TotalPagesRef)
			}
			return fmt.Sprintf("page %d", *cp.Page)
		}
	case model.PosChapter:
		if cp.Chapter != nil {
			return fmt.Sprintf("chapter %d", *cp.Chapter)
		}
	case model.PosSection:
		if cp.Section != nil && *cp.Section != "" {
			return fmt.Sprintf("section %s", *cp.Section)
		}
	case model.PosLoc:
		if cp.Loc != nil {
			return fmt.Sprintf("loc %d", *cp.Loc)
		}
	case model.PosPercent:
		if cp.Percent != nil {
			return fmt.Sprintf("%.0f%%", *cp.Percent)
		}
	case model.PosNote:
		return ""
	}
	if cp.Page != nil {
		if t != nil && t.TotalPages != nil && *t.TotalPages > 0 {
			return fmt.Sprintf("page %d / %d", *cp.Page, *t.TotalPages)
		}
		return fmt.Sprintf("page %d", *cp.Page)
	}
	if cp.Section != nil && *cp.Section != "" {
		return fmt.Sprintf("section %s", *cp.Section)
	}
	if cp.Loc != nil {
		return fmt.Sprintf("loc %d", *cp.Loc)
	}
	if cp.Percent != nil {
		return fmt.Sprintf("%.0f%%", *cp.Percent)
	}
	return ""
}

func writeListTab(w io.Writer, titles []*model.Title) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.StripEscape)
	for _, t := range titles {
		title := truncateVis(t.Title, listColTitle)
		pos := PositionSummary(t.LatestCheckpoint, t)
		if pos == "" {
			pos = "-"
		}
		pos = truncateVis(pos, listColPos)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", title, pos, string(t.Status), timefmt.TitleLineTime(t.UpdatedAt))
	}
	tw.Flush()
}

func truncateVis(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	r := []rune(s)
	if maxRunes <= 1 {
		return string(r[:maxRunes])
	}
	return string(r[:maxRunes-1]) + "..."
}

func ListTable(w io.Writer, titles []*model.Title) {
	writeListTab(w, titles)
}

func HistoryHeader(w io.Writer, t *model.Title) {
	if t != nil {
		fmt.Fprintln(w, t.Title)
	}
}

func HistoryLine(w io.Writer, cp *model.Checkpoint, t *model.Title) {
	var chunks []string
	if s := checkpointCore(cp, t); s != "" {
		chunks = append(chunks, s)
	}
	if cp.Note != nil && strings.TrimSpace(*cp.Note) != "" {
		chunks = append(chunks, "note: "+strings.TrimSpace(*cp.Note))
	}
	if pct := ProgressPercent(cp, t); pct != nil {
		chunks = append(chunks, fmt.Sprintf("%.0f%%", *pct))
	}
	chunks = append(chunks, timefmt.Relative(cp.CreatedAt))
	fmt.Fprintf(w, "  %s\n", strings.Join(chunks, " · "))
}

func SearchResults(w io.Writer, matches []search.Scored) {
	for _, m := range matches {
		if m.Title != nil {
			fmt.Fprintf(w, "%s\n", m.Title.Title)
		}
	}
}

func Ambiguous(w io.Writer, err search.AmbiguousError) {
	fmt.Fprintf(w, "dogear: several titles match %q:\n", err.Query)
	for i, m := range err.Matches {
		if m.Title == nil {
			continue
		}
		fmt.Fprintf(w, "  %d. %s\n", i+1, m.Title.Title)
	}
	fmt.Fprintf(w, "dogear: re-run with a clearer title or use list / search.\n")
}

func Stats(w io.Writer, active, finished, checkpoints int, recentTitle, staleTitle string, avgDelta *float64) {
	fmt.Fprintf(w, "active titles:     %d\n", active)
	fmt.Fprintf(w, "finished titles:   %d\n", finished)
	fmt.Fprintf(w, "checkpoints:       %d\n", checkpoints)
	if recentTitle != "" {
		fmt.Fprintf(w, "last updated:      %s\n", recentTitle)
	}
	if staleTitle != "" {
		fmt.Fprintf(w, "stalest active:    %s\n", staleTitle)
	}
	if avgDelta != nil {
		fmt.Fprintf(w, "avg page step:     %.1f\n", *avgDelta)
	}
}

func DoctorOK(w io.Writer) {
	fmt.Fprintln(w, "dogear: database looks consistent.")
}

func DoctorIssue(w io.Writer, msg string) {
	fmt.Fprintln(w, msg)
}
