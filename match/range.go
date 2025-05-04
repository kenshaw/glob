package match

import (
	"fmt"
	"unicode/utf8"

	"github.com/kenshaw/glob/debug"
)

type RangeMatcher struct {
	Lo, Hi rune
	Not    bool
}

func NewRange(lo, hi rune, not bool) RangeMatcher {
	return RangeMatcher{lo, hi, not}
}

func (self RangeMatcher) Len() int {
	return 1
}

func (self RangeMatcher) Size() int {
	return 1
}

func (self RangeMatcher) Match(s string) (ok bool) {
	if debug.Enabled {
		done := debug.Matching("range", s)
		defer func() { done(ok) }()
	}
	r, w := utf8.DecodeRuneInString(s)
	if len(s) > w {
		return false
	}
	inRange := r >= self.Lo && r <= self.Hi
	return inRange == !self.Not
}

func (self RangeMatcher) Index(s string) (index int, segments []int) {
	if debug.Enabled {
		done := debug.Indexing("range", s)
		defer func() { done(index, segments) }()
	}
	for i, r := range s {
		if self.Not != (r >= self.Lo && r <= self.Hi) {
			return i, segmentsByRuneLength[utf8.RuneLen(r)]
		}
	}
	return -1, nil
}

func (self RangeMatcher) String() string {
	var not string
	if self.Not {
		not = "!"
	}
	return fmt.Sprintf("<range:%s[%s,%s]>", not, string(self.Lo), string(self.Hi))
}
