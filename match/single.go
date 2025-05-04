package match

import (
	"fmt"
	"unicode/utf8"

	"github.com/kenshaw/glob/runes"
)

// single represents ?
type SingleMatcher struct {
	sep []rune
}

func NewSingle(s []rune) SingleMatcher {
	return SingleMatcher{s}
}

func (s SingleMatcher) Match(v string) bool {
	r, w := utf8.DecodeRuneInString(v)
	if len(v) > w {
		return false
	}
	return runes.IndexRune(s.sep, r) == -1
}

func (s SingleMatcher) Len() int {
	return 1
}

func (s SingleMatcher) Size() int {
	return 1
}

func (s SingleMatcher) Index(v string) (int, []int) {
	for i, r := range v {
		if runes.IndexRune(s.sep, r) == -1 {
			return i, segmentsByRuneLength[utf8.RuneLen(r)]
		}
	}
	return -1, nil
}

func (s SingleMatcher) String() string {
	if len(s.sep) == 0 {
		return "<single>"
	}
	return fmt.Sprintf("<single:![%s]>", string(s.sep))
}
