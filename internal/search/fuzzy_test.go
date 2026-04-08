package search

import (
	"errors"
	"testing"

	"dogear/internal/model"
)

func TestScoreExact(t *testing.T) {
	if Score("hobbit", "The Hobbit") < 0.9 {
		t.Fatalf("expected high score for substring-ish match")
	}
}

func TestBestSingleTitleExact(t *testing.T) {
	titles := []*model.Title{
		{ID: 1, Title: "The Hobbit"},
		{ID: 2, Title: "Distributed Systems Notes"},
	}
	got, err := BestSingleTitle("The Hobbit", titles)
	if err != nil || got.ID != 1 {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestBestSingleTitleFuzzy(t *testing.T) {
	titles := []*model.Title{{ID: 1, Title: "The Hobbit"}}
	got, err := BestSingleTitle("hobit", titles)
	if err != nil || got.ID != 1 {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestBestSingleTitleNoMatch(t *testing.T) {
	titles := []*model.Title{{ID: 1, Title: "The Hobbit"}}
	_, err := BestSingleTitle("quantum field theory primer", titles)
	if !errors.Is(err, ErrNoMatch) {
		t.Fatalf("expected ErrNoMatch, got %v", err)
	}
}

func TestAmbiguous(t *testing.T) {
	titles := []*model.Title{
		{ID: 1, Title: "Notes on Go"},
		{ID: 2, Title: "Notes on Rust"},
	}
	_, err := BestSingleTitle("Notes on", titles)
	if err == nil {
		t.Fatal("expected ambiguous")
	}
	var a AmbiguousError
	if !errors.As(err, &a) || len(a.Matches) < 2 {
		t.Fatalf("expected AmbiguousError, got %v", err)
	}
}
