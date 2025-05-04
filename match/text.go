package match

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// TextMatcher represents raw string to match
type TextMatcher struct {
	s     string
	runes int
	bytes int
	seg   []int
}

func NewText(s string) TextMatcher {
	return TextMatcher{
		s:     s,
		runes: utf8.RuneCountInString(s),
		bytes: len(s),
		seg:   []int{len(s)},
	}
}

func (t TextMatcher) Match(s string) bool {
	return t.s == s
}

func (t TextMatcher) Index(s string) (int, []int) {
	i := strings.Index(s, t.s)
	if i == -1 {
		return -1, nil
	}
	return i, t.seg
}

func (t TextMatcher) Len() int {
	return t.runes
}

func (t TextMatcher) BytesCount() int {
	return t.bytes
}

func (t TextMatcher) Size() int {
	return t.runes
}

func (t TextMatcher) String() string {
	return fmt.Sprintf("<text:`%v`>", t.s)
}
