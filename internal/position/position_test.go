package position

import (
	"testing"

	"dogear/internal/model"
)

func TestBuildInputPage(t *testing.T) {
	p := 12
	in, err := BuildInput(&p, nil, nil, nil, nil, nil)
	if err != nil || in.PositionType != model.PosPage {
		t.Fatalf("%+v %v", in, err)
	}
}

func TestRejectBadPercent(t *testing.T) {
	p := 101.0
	_, err := BuildInput(nil, nil, nil, nil, &p, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
