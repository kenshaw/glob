package match

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/kenshaw/glob/runes"
)

type SuffixAnyMatcher struct {
	s   string
	sep []rune
	n   int
}

func NewSuffixAny(s string, sep []rune) SuffixAnyMatcher {
	return SuffixAnyMatcher{s, sep, utf8.RuneCountInString(s)}
}

func (s SuffixAnyMatcher) Index(v string) (int, []int) {
	idx := strings.Index(v, s.s)
	if idx == -1 {
		return -1, nil
	}
	i := runes.LastIndexAnyRune(v[:idx], s.sep) + 1
	return i, []int{idx + len(s.s) - i}
}

func (s SuffixAnyMatcher) Len() int {
	return s.n
}

func (s SuffixAnyMatcher) Match(v string) bool {
	if !strings.HasSuffix(v, s.s) {
		return false
	}
	return runes.IndexAnyRune(v[:len(v)-len(s.s)], s.sep) == -1
}

func (s SuffixAnyMatcher) String() string {
	return fmt.Sprintf("<suffix_any:![%s]%s>", string(s.sep), s.s)
}
