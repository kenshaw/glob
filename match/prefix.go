package match

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type PrefixMatcher struct {
	s string
	n int
}

func NewPrefix(p string) PrefixMatcher {
	return PrefixMatcher{
		s: p,
		n: utf8.RuneCountInString(p),
	}
}

func (p PrefixMatcher) Index(s string) (int, []int) {
	idx := strings.Index(s, p.s)
	if idx == -1 {
		return -1, nil
	}
	length := len(p.s)
	var sub string
	if len(s) > idx+length {
		sub = s[idx+length:]
	} else {
		sub = ""
	}
	segments := acquireSegments(len(sub) + 1)
	segments = append(segments, length)
	for i, r := range sub {
		segments = append(segments, length+i+utf8.RuneLen(r))
	}
	return idx, segments
}

func (p PrefixMatcher) Len() int {
	return p.n
}

func (p PrefixMatcher) Match(s string) bool {
	return strings.HasPrefix(s, p.s)
}

func (p PrefixMatcher) String() string {
	return fmt.Sprintf("<prefix:%s>", p.s)
}
