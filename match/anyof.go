package match

import (
	"fmt"

	"github.com/kenshaw/glob/debug"
)

type AnyOfMatcher struct {
	v []Matcher
	n int
}

func NewAnyOf(v ...Matcher) Matcher {
	// determine minimum
	var minimum int
	for i, m := range v {
		n := m.Len()
		if i == 0 || n < minimum {
			minimum = n
		}
	}
	a := AnyOfMatcher{v, minimum}
	if mis, ok := MatchIndexers(v); ok {
		x := IndexedAnyOfMatcher{a, mis}
		if msz, ok := MatchIndexSizers(v); ok {
			sz := -1
			for _, m := range msz {
				n := m.Size()
				if sz == -1 {
					sz = n
				} else if sz != n {
					sz = -1
					break
				}
			}
			if sz != -1 {
				return IndexedSizedAnyOfMatcher{x, sz}
			}
		}
		return x
	}
	return a
}

func MustIndexedAnyOf(v ...Matcher) MatchIndexer {
	return NewAnyOf(v...).(MatchIndexer)
}

func MustIndexedSizedAnyOf(v ...Matcher) MatchIndexSizer {
	return NewAnyOf(v...).(MatchIndexSizer)
}

func (a AnyOfMatcher) Match(s string) (ok bool) {
	if debug.Enabled {
		done := debug.Matching("any_of", s)
		defer func() {
			done(ok)
		}()
	}
	for _, matcher := range a.v {
		if matcher.Match(s) {
			return true
		}
	}
	return false
}

func (a AnyOfMatcher) Len() (n int) {
	return a.n
}

func (a AnyOfMatcher) Content(f func(Matcher)) {
	for _, m := range a.v {
		f(m)
	}
}

// String satisfies the [fmt.Stringer] interface.
func (a AnyOfMatcher) String() string {
	return fmt.Sprintf("<any_of:[%s]>", Matchers(a.v))
}

type IndexedAnyOfMatcher struct {
	AnyOfMatcher
	v []MatchIndexer
}

func (a IndexedAnyOfMatcher) Index(s string) (index int, segments []int) {
	if debug.Enabled {
		done := debug.Indexing("any_of", s)
		defer func() {
			done(index, segments)
		}()
	}
	index = -1
	segments = acquireSegments(len(s))
	for _, matcher := range a.v {
		if debug.Enabled {
			debug.Logf("indexing: any_of: trying %s", matcher)
		}
		i, seg := matcher.Index(s)
		if i == -1 {
			continue
		}
		if index == -1 || i < index {
			index = i
			segments = append(segments[:0], seg...)
			continue
		}
		if i > index {
			continue
		}
		// here i == index
		segments = appendMerge(segments, seg)
	}
	if index == -1 {
		releaseSegments(segments)
		return -1, nil
	}
	return index, segments
}

func (a IndexedAnyOfMatcher) String() string {
	return fmt.Sprintf("<indexed_any_of:[%s]>", a.v)
}

type IndexedSizedAnyOfMatcher struct {
	IndexedAnyOfMatcher
	runes int
}

func (a IndexedSizedAnyOfMatcher) Size() int {
	return a.runes
}
