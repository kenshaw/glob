package match

import (
	"fmt"
	"unicode/utf8"
)

type MinMatcher struct {
	n int
}

func NewMin(n int) MinMatcher {
	return MinMatcher{n}
}

func (m MinMatcher) Match(s string) bool {
	var n int
	for range s {
		n += 1
		if n >= m.n {
			return true
		}
	}
	return false
}

func (m MinMatcher) Index(s string) (int, []int) {
	var count int
	c := len(s) - m.n + 1
	if c <= 0 {
		return -1, nil
	}
	segments := acquireSegments(c)
	for i, r := range s {
		count++
		if count >= m.n {
			segments = append(segments, i+utf8.RuneLen(r))
		}
	}
	if len(segments) == 0 {
		return -1, nil
	}
	return 0, segments
}

func (m MinMatcher) Len() int {
	return m.n
}

func (m MinMatcher) String() string {
	return fmt.Sprintf("<min:%d>", m.n)
}
