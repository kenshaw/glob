package match

import (
	"fmt"
	"strings"
)

// todo common table of rune's length

type Matcher interface {
	Match(string) bool
	Len() int
}

type Matchers []Matcher

// String satisfies the [fmt.Stringer] interface.
func (matchers Matchers) String() string {
	var s []string
	for _, matcher := range matchers {
		s = append(s, fmt.Sprint(matcher))
	}
	return fmt.Sprintf("%s", strings.Join(s, ","))
}

type Indexer interface {
	Index(string) (int, []int)
}

type Sizer interface {
	Size() int
}

type MatchSizer interface {
	Matcher
	Sizer
}

type Container interface {
	Content(func(Matcher))
}

type MatchIndexer interface {
	Matcher
	Indexer
}

func MatchIndexers(v []Matcher) ([]MatchIndexer, bool) {
	for _, m := range v {
		if _, ok := m.(MatchIndexer); !ok {
			return nil, false
		}
	}
	r := make([]MatchIndexer, len(v))
	for i := range r {
		r[i] = v[i].(MatchIndexer)
	}
	return r, true
}

type MatchIndexSizer interface {
	Matcher
	Indexer
	Sizer
}

func MatchIndexSizers(v []Matcher) ([]MatchIndexSizer, bool) {
	for _, m := range v {
		if _, ok := m.(MatchIndexSizer); !ok {
			return nil, false
		}
	}
	r := make([]MatchIndexSizer, len(v))
	for i := range r {
		r[i] = v[i].(MatchIndexSizer)
	}
	return r, true
}

// appendMerge merges and sorts given already SORTED and UNIQUE segments.
func appendMerge(target, sub []int) []int {
	nt, ns := len(target), len(sub)
	v := make([]int, 0, nt+ns)
	for i, j := 0, 0; i < nt || j < ns; {
		if i >= nt {
			v = append(v, sub[j:]...)
			break
		}
		if j >= ns {
			v = append(v, target[i:]...)
			break
		}
		t, s := target[i], sub[j]
		switch {
		case t == s:
			v = append(v, t)
			i++
			j++
		case t < s:
			v = append(v, t)
			i++
		case s < t:
			v = append(v, s)
			j++
		}
	}
	return append(target[:0], v...)
}
