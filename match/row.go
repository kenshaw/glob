package match

import (
	"fmt"
	"unicode/utf8"

	"github.com/kenshaw/glob/debug"
	"github.com/kenshaw/glob/runes"
)

type RowMatcher struct {
	ms    []MatchIndexSizer
	runes int
	seg   []int
}

func NewRow(ms []MatchIndexSizer) RowMatcher {
	var r int
	for _, m := range ms {
		r += m.Size()
	}
	return RowMatcher{
		ms:    ms,
		runes: r,
		seg:   []int{r},
	}
}

func (r RowMatcher) Match(s string) (ok bool) {
	if debug.Enabled {
		done := debug.Matching("row", s)
		defer func() { done(ok) }()
	}
	if !runes.ExactlyRunesCount(s, r.runes) {
		return false
	}
	return r.matchAll(s)
}

func (r RowMatcher) Len() int {
	return r.runes
}

func (r RowMatcher) Size() int {
	return r.runes
}

func (r RowMatcher) Index(s string) (index int, segments []int) {
	if debug.Enabled {
		done := debug.Indexing("row", s)
		debug.Logf("row: %d vs %d", len(s), r.runes)
		defer func() { done(index, segments) }()
	}
	for j := 0; j <= len(s)-r.runes; { // NOTE: using len() here to avoid counting runes.
		i, _ := r.ms[0].Index(s[j:])
		if i == -1 {
			return -1, nil
		}
		if r.matchAll(s[i:]) {
			return j + i, r.seg
		}
		_, x := utf8.DecodeRuneInString(s[i:])
		j += x
	}
	return -1, nil
}

func (r RowMatcher) Content(cb func(Matcher)) {
	for _, m := range r.ms {
		cb(m)
	}
}

func (r RowMatcher) String() string {
	return fmt.Sprintf("<row_%d:%s>", r.runes, r.ms)
}

func (r RowMatcher) matchAll(s string) bool {
	var i int
	for _, m := range r.ms {
		n := m.Size()
		sub := runes.Head(s[i:], n)
		if !m.Match(sub) {
			return false
		}
		i += len(sub)
	}
	return true
}
