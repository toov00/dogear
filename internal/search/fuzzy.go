package search

import (
	"errors"
	"math"
	"sort"
	"strings"
	"unicode"

	"dogear/internal/model"
)

const defaultThreshold = 0.32

const ambiguityGap = 0.055

const maxAmbiguityChoices = 8

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	row := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		row[j] = j
	}
	for i := 1; i <= la; i++ {
		row[0] = i
		prev := i - 1
		for j := 1; j <= lb; j++ {
			cur := row[j]
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			del := row[j] + 1
			ins := row[j-1] + 1
			sub := prev + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			prev = cur
			row[j] = m
		}
	}
	return row[lb]
}

func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func ratioDistance(q, c string) float64 {
	rq, rc := []rune(q), []rune(c)
	if len(rq) == 0 || len(rc) == 0 {
		return 0
	}
	maxLen := len(rq)
	if len(rc) > maxLen {
		maxLen = len(rc)
	}
	dist := levenshtein(q, c)
	return 1 - float64(dist)/float64(maxLen)
}

func tokenCoverage(query, candidate string) float64 {
	q := normalize(query)
	c := normalize(candidate)
	if q == "" {
		return 0
	}
	words := strings.Fields(q)
	if len(words) == 0 {
		return 0
	}
	hit := 0
	for _, w := range words {
		if len(w) < 2 {
			continue
		}
		if strings.Contains(c, w) {
			hit++
		}
	}
	if hit == 0 && len(words) == 1 && len(words[0]) >= 2 && strings.Contains(c, words[0]) {
		return 1
	}
	if hit == 0 {
		return 0
	}
	return float64(hit) / float64(len(words))
}

func prefixBoost(query, candidate string) float64 {
	q := normalize(query)
	c := normalize(candidate)
	if q == "" || c == "" {
		return 0
	}
	if strings.HasPrefix(c, q) {
		return 0.06
	}
	for _, suf := range []string{" ", ".", ":", "-", "'"} {
		if strings.HasPrefix(c, q+suf) {
			return 0.04
		}
	}
	return 0
}

func Score(query, candidate string) float64 {
	q := normalize(query)
	c := normalize(candidate)
	if q == "" {
		return 0
	}
	if c == q {
		return 1
	}
	var base float64
	if strings.Contains(c, q) {
		base = 0.93
	} else {
		base = ratioDistance(q, c)
	}
	tok := tokenCoverage(query, candidate)
	combined := math.Max(base, tok*0.97)
	if tok >= 0.99 && len(strings.Fields(q)) > 1 {
		combined = math.Max(combined, 0.88)
	}
	combined += prefixBoost(query, candidate)
	if combined > 1 {
		combined = 1
	}
	return combined
}

func thresholdFor(query string) float64 {
	q := strings.TrimSpace(query)
	if len([]rune(q)) >= 12 {
		return defaultThreshold - 0.03
	}
	return defaultThreshold
}

type Scored struct {
	Title *model.Title
	Score float64
}

func MatchTitles(query string, titles []*model.Title, threshold float64) []Scored {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil
	}
	th := threshold
	if th <= 0 {
		th = thresholdFor(q)
	}
	var out []Scored
	for _, t := range titles {
		if t == nil {
			continue
		}
		s := Score(q, t.Title)
		if s >= th {
			out = append(out, Scored{Title: t, Score: s})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return strings.ToLower(out[i].Title.Title) < strings.ToLower(out[j].Title.Title)
	})
	return out
}

var ErrNoMatch = errors.New("no matching title found")

type AmbiguousError struct {
	Query   string
	Matches []Scored
}

func (e AmbiguousError) Error() string {
	return "multiple titles match; please be more specific"
}

func BestSingleTitle(query string, titles []*model.Title) (*model.Title, error) {
	th := thresholdFor(query)
	matches := MatchTitles(query, titles, th)
	if len(matches) == 0 {
		return nil, ErrNoMatch
	}
	if len(matches) == 1 {
		return matches[0].Title, nil
	}
	top := matches[0].Score
	sec := matches[1].Score
	if sec >= top-ambiguityGap {
		return nil, AmbiguousError{Query: query, Matches: trimAmbiguityList(matches)}
	}
	if top < 0.58 && sec >= top-0.1 {
		return nil, AmbiguousError{Query: query, Matches: trimAmbiguityList(matches)}
	}
	return matches[0].Title, nil
}

func trimAmbiguityList(m []Scored) []Scored {
	if len(m) <= maxAmbiguityChoices {
		return m
	}
	return m[:maxAmbiguityChoices]
}

func IsNoMatch(err error) bool {
	return errors.Is(err, ErrNoMatch)
}

func IsAmbiguous(err error) (*AmbiguousError, bool) {
	var a AmbiguousError
	if errors.As(err, &a) {
		return &a, true
	}
	return nil, false
}
