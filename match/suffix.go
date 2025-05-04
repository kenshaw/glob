package match

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type SuffixMatcher struct {
	s string
	n int
}

func NewSuffix(s string) SuffixMatcher {
	return SuffixMatcher{s, utf8.RuneCountInString(s)}
}

func (s SuffixMatcher) Len() int {
	return s.n
}

func (s SuffixMatcher) Match(v string) bool {
	return strings.HasSuffix(v, s.s)
}

func (s SuffixMatcher) Index(v string) (int, []int) {
	idx := strings.Index(v, s.s)
	if idx == -1 {
		return -1, nil
	}
	return 0, []int{idx + len(s.s)}
}

func (s SuffixMatcher) String() string {
	return fmt.Sprintf("<suffix:%s>", s.s)
}
