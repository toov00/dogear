package output

import (
	"encoding/json"
	"io"
	"time"

	"dogear/internal/model"
	"dogear/internal/search"
)

type jsonWhere struct {
	Title           string   `json:"title"`
	AuthorOrSource  string   `json:"author_or_source,omitempty"`
	Format          string   `json:"format,omitempty"`
	Status          string   `json:"status"`
	Tags            []string `json:"tags,omitempty"`
	Position        string   `json:"position,omitempty"`
	ProgressPercent *float64 `json:"progress_percent,omitempty"`
	UpdatedAt       string   `json:"updated_at"`
	Note            string   `json:"note,omitempty"`
}

func WriteWhereJSON(w io.Writer, t *model.Title) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if t == nil {
		return enc.Encode(jsonWhere{})
	}
	var note string
	if t.LatestCheckpoint != nil && t.LatestCheckpoint.Note != nil {
		note = *t.LatestCheckpoint.Note
	}
	var pct *float64
	if t.LatestCheckpoint != nil {
		pct = ProgressPercent(t.LatestCheckpoint, t)
	}
	out := jsonWhere{
		Title:           t.Title,
		AuthorOrSource:  t.AuthorOrSource,
		Format:          t.Format,
		Status:          string(t.Status),
		Tags:            append([]string(nil), t.Tags...),
		Position:        PositionSummary(t.LatestCheckpoint, t),
		ProgressPercent: pct,
		UpdatedAt:       t.UpdatedAt.UTC().Format(time.RFC3339),
		Note:            note,
	}
	return enc.Encode(out)
}

type jsonListRow struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	Tags      []string `json:"tags,omitempty"`
	Position  string   `json:"position,omitempty"`
	UpdatedAt string   `json:"updated_at"`
}

func WriteListJSON(w io.Writer, titles []*model.Title) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	rows := make([]jsonListRow, 0, len(titles))
	for _, t := range titles {
		if t == nil {
			continue
		}
		rows = append(rows, jsonListRow{
			ID:        t.ID,
			Title:     t.Title,
			Status:    string(t.Status),
			Tags:      append([]string(nil), t.Tags...),
			Position:  PositionSummary(t.LatestCheckpoint, t),
			UpdatedAt: t.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	return enc.Encode(rows)
}

type jsonHistoryOut struct {
	Title       string           `json:"title"`
	Checkpoints []jsonHistoryRow `json:"checkpoints"`
}

type jsonHistoryRow struct {
	Position        string   `json:"position,omitempty"`
	ProgressPercent *float64 `json:"progress_percent,omitempty"`
	Note            string   `json:"note,omitempty"`
	CreatedAt       string   `json:"created_at"`
}

func WriteHistoryJSON(w io.Writer, t *model.Title, cps []*model.Checkpoint) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	out := jsonHistoryOut{Title: t.Title}
	for _, cp := range cps {
		if cp == nil {
			continue
		}
		row := jsonHistoryRow{
			Position:        checkpointCore(cp, t),
			ProgressPercent: ProgressPercent(cp, t),
			CreatedAt:       cp.CreatedAt.UTC().Format(time.RFC3339),
		}
		if cp.Note != nil {
			row.Note = *cp.Note
		}
		out.Checkpoints = append(out.Checkpoints, row)
	}
	return enc.Encode(out)
}

func WriteSearchJSON(w io.Writer, matches []search.Scored) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	type row struct {
		ID    int64   `json:"id"`
		Title string  `json:"title"`
		Score float64 `json:"score"`
	}
	var rows []row
	for _, m := range matches {
		if m.Title == nil {
			continue
		}
		rows = append(rows, row{ID: m.Title.ID, Title: m.Title.Title, Score: roundScore(m.Score)})
	}
	return enc.Encode(rows)
}

func roundScore(s float64) float64 {
	return float64(int(s*1000+0.5)) / 1000
}

type jsonStats struct {
	ActiveTitles    int      `json:"active_titles"`
	FinishedTitles  int      `json:"finished_titles"`
	Checkpoints     int      `json:"checkpoints"`
	LastUpdated     string   `json:"last_updated_title,omitempty"`
	StalestActive   string   `json:"stalest_active,omitempty"`
	AvgPageStep     *float64 `json:"avg_page_step,omitempty"`
}

func WriteStatsJSON(w io.Writer, active, finished, checkpoints int, recentTitle, staleTitle string, avgDelta *float64) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	s := jsonStats{
		ActiveTitles:    active,
		FinishedTitles:  finished,
		Checkpoints:     checkpoints,
		LastUpdated:     recentTitle,
		StalestActive:   staleTitle,
		AvgPageStep:     avgDelta,
	}
	return enc.Encode(s)
}
