package match

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/kenshaw/glob/runes"
)

type PrefixAnyMatcher struct {
	s   string
	sep []rune
	n   int
}

func NewPrefixAny(s string, sep []rune) PrefixAnyMatcher {
	return PrefixAnyMatcher{s, sep, utf8.RuneCountInString(s)}
}

func (p PrefixAnyMatcher) Index(s string) (int, []int) {
	idx := strings.Index(s, p.s)
	if idx == -1 {
		return -1, nil
	}
	n := len(p.s)
	sub := s[idx+n:]
	i := runes.IndexAnyRune(sub, p.sep)
	if i > -1 {
		sub = sub[:i]
	}
	seg := acquireSegments(len(sub) + 1)
	seg = append(seg, n)
	for i, r := range sub {
		seg = append(seg, n+i+utf8.RuneLen(r))
	}
	return idx, seg
}

func (p PrefixAnyMatcher) Len() int {
	return p.n
}

func (p PrefixAnyMatcher) Match(s string) bool {
	if !strings.HasPrefix(s, p.s) {
		return false
	}
	return runes.IndexAnyRune(s[len(p.s):], p.sep) == -1
}

func (p PrefixAnyMatcher) String() string {
	return fmt.Sprintf("<prefix_any:%s![%s]>", p.s, string(p.sep))
}
