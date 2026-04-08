package model

import "time"

type TitleStatus string

const (
	StatusActive   TitleStatus = "active"
	StatusFinished TitleStatus = "finished"
)

type Title struct {
	ID              int64
	Title           string
	AuthorOrSource  string
	Format          string
	TotalPages      *int
	TotalChapters   *int
	Status          TitleStatus
	Tags            []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FinishedAt      *time.Time
	LatestCheckpoint *Checkpoint
}

type PositionType string

const (
	PosPage    PositionType = "page"
	PosChapter PositionType = "chapter"
	PosSection PositionType = "section"
	PosLoc     PositionType = "loc"
	PosPercent PositionType = "percent"
	PosNote    PositionType = "note"
)

type Checkpoint struct {
	ID            int64
	TitleID       int64
	PositionType  PositionType
	Page          *int
	Chapter       *int
	Section       *string
	Loc           *int
	Percent       *float64
	Note          *string
	CreatedAt     time.Time
	TotalPagesRef *int
}

type ExportDoc struct {
	Version int         `json:"version"`
	Titles  []ExportTitle `json:"titles"`
}

type ExportTitle struct {
	ID             int64              `json:"id"`
	Title          string             `json:"title"`
	AuthorOrSource string             `json:"author_or_source,omitempty"`
	Format         string             `json:"format,omitempty"`
	TotalPages     *int               `json:"total_pages,omitempty"`
	TotalChapters  *int               `json:"total_chapters,omitempty"`
	Status         string             `json:"status"`
	Tags           []string           `json:"tags,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	FinishedAt     *time.Time         `json:"finished_at,omitempty"`
	Checkpoints    []ExportCheckpoint `json:"checkpoints"`
}

type ExportCheckpoint struct {
	ID           int64     `json:"id"`
	PositionType string    `json:"position_type"`
	Page         *int      `json:"page,omitempty"`
	Chapter      *int      `json:"chapter,omitempty"`
	Section      *string   `json:"section,omitempty"`
	Loc          *int      `json:"loc,omitempty"`
	Percent      *float64  `json:"percent,omitempty"`
	Note         *string   `json:"note,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type TitleInput struct {
	Title          string
	AuthorOrSource string
	Format         string
	TotalPages     *int
	TotalChapters  *int
	Tags           []string
}

type CheckpointInput struct {
	PositionType PositionType
	Page           *int
	Chapter        *int
	Section        *string
	Loc            *int
	Percent        *float64
	Note           *string
}
