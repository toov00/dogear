package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"dogear/internal/model"
)

type Store struct {
	db *sql.DB
}

func New(database *sql.DB) *Store {
	return &Store{db: database}
}

func (r *Store) HasDuplicateActive(ctx context.Context, title, format string) (bool, error) {
	const q = `SELECT 1 FROM titles WHERE lower(title) = lower(?) AND lower(format) = lower(?) AND status = 'active' LIMIT 1`
	var x int
	err := r.db.QueryRowContext(ctx, q, title, format).Scan(&x)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *Store) InsertTitle(ctx context.Context, in model.TitleInput, now time.Time) (int64, error) {
	const q = `INSERT INTO titles (title, author_or_source, format, total_pages, total_chapters, status, created_at, updated_at, finished_at)
		VALUES (?,?,?,?,?,'active',?,?,NULL)`
	var tp, tc sql.NullInt64
	if in.TotalPages != nil {
		tp = sql.NullInt64{Int64: int64(*in.TotalPages), Valid: true}
	}
	if in.TotalChapters != nil {
		tc = sql.NullInt64{Int64: int64(*in.TotalChapters), Valid: true}
	}
	res, err := r.db.ExecContext(ctx, q,
		in.Title,
		in.AuthorOrSource,
		in.Format,
		tp,
		tc,
		now.UTC().Format(time.RFC3339),
		now.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Store) ReplaceTags(ctx context.Context, titleID int64, tags []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM title_tags WHERE title_id = ?`, titleID); err != nil {
		return err
	}
	for _, t := range tags {
		t = strings.TrimSpace(strings.ToLower(t))
		if t == "" {
			continue
		}
		if _, err := r.db.ExecContext(ctx, `INSERT INTO title_tags (title_id, tag) VALUES (?,?)`, titleID, t); err != nil {
			return err
		}
	}
	return nil
}

func (r *Store) AddTag(ctx context.Context, titleID int64, tag string) error {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" {
		return fmt.Errorf("tag is empty")
	}
	_, err := r.db.ExecContext(ctx, `INSERT OR IGNORE INTO title_tags (title_id, tag) VALUES (?,?)`, titleID, tag)
	return err
}

func (r *Store) RemoveTag(ctx context.Context, titleID int64, tag string) error {
	tag = strings.TrimSpace(strings.ToLower(tag))
	_, err := r.db.ExecContext(ctx, `DELETE FROM title_tags WHERE title_id = ? AND tag = ?`, titleID, tag)
	return err
}

func (r *Store) InsertCheckpoint(ctx context.Context, titleID int64, in model.CheckpointInput, at time.Time) (int64, error) {
	pt := string(in.PositionType)
	var page, chapter, loc sql.NullInt64
	var section, note sql.NullString
	var pct sql.NullFloat64
	if in.Page != nil {
		page = sql.NullInt64{Int64: int64(*in.Page), Valid: true}
	}
	if in.Chapter != nil {
		chapter = sql.NullInt64{Int64: int64(*in.Chapter), Valid: true}
	}
	if in.Section != nil {
		section = sql.NullString{String: *in.Section, Valid: true}
	}
	if in.Loc != nil {
		loc = sql.NullInt64{Int64: int64(*in.Loc), Valid: true}
	}
	if in.Percent != nil {
		pct = sql.NullFloat64{Float64: *in.Percent, Valid: true}
	}
	if in.Note != nil {
		note = sql.NullString{String: *in.Note, Valid: true}
	}
	const q = `INSERT INTO checkpoints (title_id, position_type, page, chapter, section, loc, percent, note, created_at)
		VALUES (?,?,?,?,?,?,?,?,?)`
	res, err := r.db.ExecContext(ctx, q, titleID, pt, page, chapter, section, loc, pct, note, at.UTC().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	if _, err := r.db.ExecContext(ctx, `UPDATE titles SET updated_at = ? WHERE id = ?`, at.UTC().Format(time.RFC3339), titleID); err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Store) LoadAllTitles(ctx context.Context) ([]*model.Title, error) {
	const q = `SELECT id, title, author_or_source, format, total_pages, total_chapters, status, created_at, updated_at, finished_at FROM titles`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Title
	for rows.Next() {
		t, err := scanTitleRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func scanTitleRow(rows *sql.Rows) (*model.Title, error) {
	var (
		id                                                          int64
		title, author, format, status, createdS, updatedS, fin sql.NullString
		totalP, totalC                                              sql.NullInt64
	)
	if err := rows.Scan(&id, &title, &author, &format, &totalP, &totalC, &status, &createdS, &updatedS, &fin); err != nil {
		return nil, err
	}
	t := &model.Title{
		ID:             id,
		Title:          title.String,
		AuthorOrSource: author.String,
		Format:         format.String,
		Status:         model.TitleStatus(status.String),
	}
	if totalP.Valid {
		v := int(totalP.Int64)
		t.TotalPages = &v
	}
	if totalC.Valid {
		v := int(totalC.Int64)
		t.TotalChapters = &v
	}
	var err error
	t.CreatedAt, err = time.Parse(time.RFC3339, createdS.String)
	if err != nil {
		return nil, err
	}
	t.UpdatedAt, err = time.Parse(time.RFC3339, updatedS.String)
	if err != nil {
		return nil, err
	}
	if fin.Valid && fin.String != "" {
		ft, e := time.Parse(time.RFC3339, fin.String)
		if e != nil {
			return nil, e
		}
		t.FinishedAt = &ft
	}
	return t, nil
}

func (r *Store) loadTagsFor(ctx context.Context, ids []int64) (map[int64][]string, error) {
	if len(ids) == 0 {
		return map[int64][]string{}, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i := range ids {
		placeholders[i] = "?"
		args[i] = ids[i]
	}
	q := `SELECT title_id, tag FROM title_tags WHERE title_id IN (` + strings.Join(placeholders, ",") + `) ORDER BY tag`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[int64][]string{}
	for rows.Next() {
		var id int64
		var tag string
		if err := rows.Scan(&id, &tag); err != nil {
			return nil, err
		}
		m[id] = append(m[id], tag)
	}
	return m, rows.Err()
}

func (r *Store) LoadTitleFull(ctx context.Context, id int64) (*model.Title, error) {
	const q = `SELECT id, title, author_or_source, format, total_pages, total_chapters, status, created_at, updated_at, finished_at FROM titles WHERE id = ?`
	row := r.db.QueryRowContext(ctx, q, id)
	var (
		tid                                int64
		title, author, format, status      string
		createdS, updatedS                 string
		totalP, totalC                     sql.NullInt64
		fin                                sql.NullString
	)
	err := row.Scan(&tid, &title, &author, &format, &totalP, &totalC, &status, &createdS, &updatedS, &fin)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t := &model.Title{
		ID:             tid,
		Title:          title,
		AuthorOrSource: author,
		Format:         format,
		Status:         model.TitleStatus(status),
	}
	if totalP.Valid {
		v := int(totalP.Int64)
		t.TotalPages = &v
	}
	if totalC.Valid {
		v := int(totalC.Int64)
		t.TotalChapters = &v
	}
	t.CreatedAt, err = time.Parse(time.RFC3339, createdS)
	if err != nil {
		return nil, err
	}
	t.UpdatedAt, err = time.Parse(time.RFC3339, updatedS)
	if err != nil {
		return nil, err
	}
	if fin.Valid && fin.String != "" {
		ft, e := time.Parse(time.RFC3339, fin.String)
		if e != nil {
			return nil, e
		}
		t.FinishedAt = &ft
	}
	tags, err := r.loadTagsFor(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	t.Tags = tags[id]
	cp, err := r.latestCheckpoint(ctx, id)
	if err != nil {
		return nil, err
	}
	t.LatestCheckpoint = cp
	return t, nil
}

func (r *Store) latestCheckpoint(ctx context.Context, titleID int64) (*model.Checkpoint, error) {
	const q = `SELECT id, title_id, position_type, page, chapter, section, loc, percent, note, created_at FROM checkpoints WHERE title_id = ? ORDER BY created_at DESC, id DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, q, titleID)
	return scanCheckpoint(row)
}

func scanCheckpoint(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.Checkpoint, error) {
	var (
		id, tid                                                 int64
		posType, createdS                                       string
		page, chapter, loc                                      sql.NullInt64
		section, note                                           sql.NullString
		pct                                                     sql.NullFloat64
	)
	err := scanner.Scan(&id, &tid, &posType, &page, &chapter, &section, &loc, &pct, &note, &createdS)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	cp := &model.Checkpoint{
		ID:           id,
		TitleID:      tid,
		PositionType: model.PositionType(posType),
	}
	if page.Valid {
		v := int(page.Int64)
		cp.Page = &v
	}
	if chapter.Valid {
		v := int(chapter.Int64)
		cp.Chapter = &v
	}
	if section.Valid {
		s := section.String
		cp.Section = &s
	}
	if loc.Valid {
		v := int(loc.Int64)
		cp.Loc = &v
	}
	if pct.Valid {
		v := pct.Float64
		cp.Percent = &v
	}
	if note.Valid {
		n := note.String
		cp.Note = &n
	}
	cp.CreatedAt, err = time.Parse(time.RFC3339, createdS)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

func (r *Store) ListTitles(ctx context.Context, status string, tag string) ([]*model.Title, error) {
	var conds []string
	var args []interface{}
	switch status {
	case "active":
		conds = append(conds, "titles.status = 'active'")
	case "finished":
		conds = append(conds, "titles.status = 'finished'")
	}
	q := `SELECT titles.id, titles.title, titles.author_or_source, titles.format, titles.total_pages, titles.total_chapters, titles.status, titles.created_at, titles.updated_at, titles.finished_at FROM titles`
	if tag != "" {
		q += ` INNER JOIN title_tags ON title_tags.title_id = titles.id AND title_tags.tag = ?`
		args = append(args, strings.TrimSpace(strings.ToLower(tag)))
	}
	if len(conds) > 0 {
		q += ` WHERE ` + strings.Join(conds, " AND ")
	}
	order := ` ORDER BY CASE WHEN titles.status = 'active' THEN 0 ELSE 1 END, titles.updated_at DESC`
	q += order
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Title
	var ids []int64
	for rows.Next() {
		t, err := scanTitleRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
		ids = append(ids, t.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	tm, err := r.loadTagsFor(ctx, ids)
	if err != nil {
		return nil, err
	}
	latest, err := r.loadLatestCheckpoints(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, t := range out {
		t.Tags = tm[t.ID]
		t.LatestCheckpoint = latest[t.ID]
	}
	return out, nil
}

func (r *Store) loadLatestCheckpoints(ctx context.Context, ids []int64) (map[int64]*model.Checkpoint, error) {
	if len(ids) == 0 {
		return map[int64]*model.Checkpoint{}, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i := range ids {
		placeholders[i] = "?"
		args[i] = ids[i]
	}
	q := `SELECT id, title_id, position_type, page, chapter, section, loc, percent, note, created_at FROM (
		SELECT id, title_id, position_type, page, chapter, section, loc, percent, note, created_at,
			row_number() OVER (PARTITION BY title_id ORDER BY created_at DESC, id DESC) AS rn
		FROM checkpoints WHERE title_id IN (` + strings.Join(placeholders, ",") + `)
	) WHERE rn = 1`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]*model.Checkpoint{}
	for rows.Next() {
		cp, err := scanCheckpoint(rows)
		if err != nil || cp == nil {
			continue
		}
		out[cp.TitleID] = cp
	}
	return out, rows.Err()
}

func (r *Store) Lately(ctx context.Context, limit int) ([]*model.Title, error) {
	if limit <= 0 {
		limit = 20
	}
	const q = `SELECT id, title, author_or_source, format, total_pages, total_chapters, status, created_at, updated_at, finished_at FROM titles WHERE status = 'active' ORDER BY updated_at DESC LIMIT ?`
	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Title
	var ids []int64
	for rows.Next() {
		t, err := scanTitleRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
		ids = append(ids, t.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	tm, err := r.loadTagsFor(ctx, ids)
	if err != nil {
		return nil, err
	}
	latest, err := r.loadLatestCheckpoints(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, t := range out {
		t.Tags = tm[t.ID]
		t.LatestCheckpoint = latest[t.ID]
	}
	return out, nil
}

func (r *Store) History(ctx context.Context, titleID int64) ([]*model.Checkpoint, error) {
	const q = `SELECT id, title_id, position_type, page, chapter, section, loc, percent, note, created_at FROM checkpoints WHERE title_id = ? ORDER BY created_at DESC, id DESC`
	rows, err := r.db.QueryContext(ctx, q, titleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Checkpoint
	for rows.Next() {
		cp, err := scanCheckpoint(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, cp)
	}
	return out, rows.Err()
}

func (r *Store) Finish(ctx context.Context, titleID int64, at time.Time) error {
	s := at.UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `UPDATE titles SET status = 'finished', finished_at = ?, updated_at = ? WHERE id = ?`, s, s, titleID)
	return err
}

func (r *Store) DeleteTitle(ctx context.Context, titleID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM titles WHERE id = ?`, titleID)
	return err
}

func (r *Store) Counts(ctx context.Context) (active, finished, checkpoints int, err error) {
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM titles WHERE status = 'active'`).Scan(&active)
	if err != nil {
		return
	}
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM titles WHERE status = 'finished'`).Scan(&finished)
	if err != nil {
		return
	}
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM checkpoints`).Scan(&checkpoints)
	return
}

func (r *Store) MostRecentlyUpdatedTitle(ctx context.Context) (string, error) {
	var title sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT title FROM titles ORDER BY updated_at DESC LIMIT 1`).Scan(&title)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return title.String, nil
}

func (r *Store) StalestActive(ctx context.Context) (string, error) {
	var title sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT title FROM titles WHERE status = 'active' ORDER BY updated_at ASC LIMIT 1`).Scan(&title)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return title.String, nil
}

func (r *Store) StaleActive(ctx context.Context, days int) ([]*model.Title, error) {
	const q = `SELECT id, title, author_or_source, format, total_pages, total_chapters, status, created_at, updated_at, finished_at FROM titles WHERE status = 'active' AND datetime(updated_at) < datetime('now', ?) ORDER BY updated_at ASC`
	rows, err := r.db.QueryContext(ctx, q, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Title
	var ids []int64
	for rows.Next() {
		t, err := scanTitleRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
		ids = append(ids, t.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	tm, err := r.loadTagsFor(ctx, ids)
	if err != nil {
		return nil, err
	}
	latest, err := r.loadLatestCheckpoints(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, t := range out {
		t.Tags = tm[t.ID]
		t.LatestCheckpoint = latest[t.ID]
	}
	return out, nil
}

func (r *Store) PageStepAverages(ctx context.Context) (*float64, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT title_id, page FROM checkpoints WHERE page IS NOT NULL ORDER BY title_id, created_at ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lastTitle int64 = -1
	var lastPage int
	var sum float64
	var n int
	for rows.Next() {
		var tid int64
		var page int
		if err := rows.Scan(&tid, &page); err != nil {
			return nil, err
		}
		if tid == lastTitle && page > 0 && lastPage > 0 {
			delta := float64(page - lastPage)
			if delta > 0 {
				sum += delta
				n++
			}
		}
		lastTitle = tid
		lastPage = page
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	avg := sum / float64(n)
	return &avg, nil
}

func (r *Store) Doctor(ctx context.Context) []string {
	var issues []string
	var c int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM checkpoints c LEFT JOIN titles t ON t.id = c.title_id WHERE t.id IS NULL`).Scan(&c)
	if c > 0 {
		issues = append(issues, fmt.Sprintf("Found %d checkpoints with missing titles.", c))
	}
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM title_tags x LEFT JOIN titles t ON t.id = x.title_id WHERE t.id IS NULL`).Scan(&c)
	if c > 0 {
		issues = append(issues, fmt.Sprintf("Found %d tag rows with missing titles.", c))
	}
	return issues
}

func (r *Store) ExportSnapshot(ctx context.Context) (*model.ExportDoc, error) {
	titles, err := r.LoadAllTitles(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(titles, func(i, j int) bool { return titles[i].ID < titles[j].ID })
	var ids []int64
	for _, t := range titles {
		ids = append(ids, t.ID)
	}
	tm, err := r.loadTagsFor(ctx, ids)
	if err != nil {
		return nil, err
	}
	doc := &model.ExportDoc{Version: 1, Titles: []model.ExportTitle{}}
	for _, t := range titles {
		t.Tags = tm[t.ID]
		cps, err := r.History(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		sort.Slice(cps, func(i, j int) bool {
			if cps[i].CreatedAt.Equal(cps[j].CreatedAt) {
				return cps[i].ID < cps[j].ID
			}
			return cps[i].CreatedAt.Before(cps[j].CreatedAt)
		})
		et := model.ExportTitle{
			ID:             t.ID,
			Title:          t.Title,
			AuthorOrSource: t.AuthorOrSource,
			Format:         t.Format,
			TotalPages:     t.TotalPages,
			TotalChapters:  t.TotalChapters,
			Status:         string(t.Status),
			Tags:           t.Tags,
			CreatedAt:      t.CreatedAt,
			UpdatedAt:      t.UpdatedAt,
			FinishedAt:     t.FinishedAt,
		}
		for _, c := range cps {
			ec := model.ExportCheckpoint{
				ID:           c.ID,
				PositionType: string(c.PositionType),
				Page:         c.Page,
				Chapter:      c.Chapter,
				Section:      c.Section,
				Loc:          c.Loc,
				Percent:      c.Percent,
				Note:         c.Note,
				CreatedAt:    c.CreatedAt,
			}
			et.Checkpoints = append(et.Checkpoints, ec)
		}
		doc.Titles = append(doc.Titles, et)
	}
	return doc, nil
}

func (r *Store) ImportReplace(ctx context.Context, doc *model.ExportDoc) error {
	if doc.Version == 0 {
		doc.Version = 1
	}
	if err := doc.Validate(); err != nil {
		return err
	}
	sort.Slice(doc.Titles, func(i, j int) bool { return doc.Titles[i].ID < doc.Titles[j].ID })
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM checkpoints`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM title_tags`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM titles`); err != nil {
		return err
	}
	_, _ = tx.ExecContext(ctx, `DELETE FROM sqlite_sequence WHERE name IN ('titles','checkpoints')`)
	for _, et := range doc.Titles {
		var tp, tc interface{}
		if et.TotalPages != nil {
			tp = *et.TotalPages
		}
		if et.TotalChapters != nil {
			tc = *et.TotalChapters
		}
		var fin interface{}
		if et.FinishedAt != nil {
			fin = et.FinishedAt.UTC().Format(time.RFC3339)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO titles (id, title, author_or_source, format, total_pages, total_chapters, status, created_at, updated_at, finished_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
			et.ID, et.Title, et.AuthorOrSource, et.Format, tp, tc, et.Status,
			et.CreatedAt.UTC().Format(time.RFC3339),
			et.UpdatedAt.UTC().Format(time.RFC3339),
			fin,
		); err != nil {
			return err
		}
		for _, tag := range et.Tags {
			tag = strings.TrimSpace(strings.ToLower(tag))
			if tag == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO title_tags (title_id, tag) VALUES (?,?)`, et.ID, tag); err != nil {
				return err
			}
		}
		for _, ec := range et.Checkpoints {
			var page, chapter, loc sql.NullInt64
			var section, note sql.NullString
			var pct sql.NullFloat64
			if ec.Page != nil {
				page = sql.NullInt64{Int64: int64(*ec.Page), Valid: true}
			}
			if ec.Chapter != nil {
				chapter = sql.NullInt64{Int64: int64(*ec.Chapter), Valid: true}
			}
			if ec.Section != nil {
				section = sql.NullString{String: *ec.Section, Valid: true}
			}
			if ec.Loc != nil {
				loc = sql.NullInt64{Int64: int64(*ec.Loc), Valid: true}
			}
			if ec.Percent != nil {
				pct = sql.NullFloat64{Float64: *ec.Percent, Valid: true}
			}
			if ec.Note != nil {
				note = sql.NullString{String: *ec.Note, Valid: true}
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO checkpoints (id, title_id, position_type, page, chapter, section, loc, percent, note, created_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
				ec.ID, et.ID, ec.PositionType, page, chapter, section, loc, pct, note, ec.CreatedAt.UTC().Format(time.RFC3339)); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}
