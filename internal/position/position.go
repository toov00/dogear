package position

import (
	"fmt"
	"strings"

	"dogear/internal/model"
)

func BuildInput(page *int, chapter *int, section *string, loc *int, percent *float64, note *string) (model.CheckpointInput, error) {
	in := model.CheckpointInput{
		Page:     page,
		Chapter:  chapter,
		Section:  section,
		Loc:      loc,
		Percent:  percent,
		Note:     note,
	}
	if err := validateFields(&in); err != nil {
		return model.CheckpointInput{}, err
	}
	if !hasAnyPosition(&in) {
		return model.CheckpointInput{}, fmt.Errorf("provide at least one of page, chapter, section, loc, percent, or note")
	}
	in.PositionType = inferType(&in)
	return in, nil
}

func BuildOptional(page *int, chapter *int, section *string, loc *int, percent *float64, note *string) (*model.CheckpointInput, error) {
	has := page != nil || chapter != nil || (section != nil && strings.TrimSpace(*section) != "") || loc != nil || percent != nil || (note != nil && strings.TrimSpace(*note) != "")
	if !has {
		return nil, nil
	}
	in, err := BuildInput(page, chapter, section, loc, percent, note)
	if err != nil {
		return nil, err
	}
	return &in, nil
}

func validateFields(in *model.CheckpointInput) error {
	if in.Page != nil && *in.Page <= 0 {
		return fmt.Errorf("page must be a positive number")
	}
	if in.Chapter != nil && *in.Chapter <= 0 {
		return fmt.Errorf("chapter must be a positive number")
	}
	if in.Loc != nil && *in.Loc <= 0 {
		return fmt.Errorf("loc must be a positive number")
	}
	if in.Percent != nil {
		p := *in.Percent
		if p < 0 || p > 100 {
			return fmt.Errorf("percent must be between 0 and 100")
		}
	}
	return nil
}

func hasAnyPosition(in *model.CheckpointInput) bool {
	if in.Page != nil {
		return true
	}
	if in.Chapter != nil {
		return true
	}
	if in.Section != nil && strings.TrimSpace(*in.Section) != "" {
		return true
	}
	if in.Loc != nil {
		return true
	}
	if in.Percent != nil {
		return true
	}
	if in.Note != nil && strings.TrimSpace(*in.Note) != "" {
		return true
	}
	return false
}

func inferType(in *model.CheckpointInput) model.PositionType {
	if in.Page != nil {
		return model.PosPage
	}
	if in.Section != nil && strings.TrimSpace(*in.Section) != "" {
		return model.PosSection
	}
	if in.Chapter != nil {
		return model.PosChapter
	}
	if in.Loc != nil {
		return model.PosLoc
	}
	if in.Percent != nil {
		return model.PosPercent
	}
	return model.PosNote
}
