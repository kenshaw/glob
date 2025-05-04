package match

import (
	"fmt"

	"github.com/kenshaw/glob/runes"
)

type AnyMatcher struct {
	sep []rune
}

func NewAny(s []rune) AnyMatcher {
	return AnyMatcher{s}
}

func (a AnyMatcher) Match(s string) bool {
	return runes.IndexAnyRune(s, a.sep) == -1
}

func (a AnyMatcher) Index(s string) (int, []int) {
	found := runes.IndexAnyRune(s, a.sep)
	switch found {
	case -1:
	case 0:
		return 0, segments0
	default:
		s = s[:found]
	}
	segments := acquireSegments(len(s))
	for i := range s {
		segments = append(segments, i)
	}
	segments = append(segments, len(s))
	return 0, segments
}

func (a AnyMatcher) Len() int {
	return 0
}

func (a AnyMatcher) String() string {
	return fmt.Sprintf("<any:![%s]>", string(a.sep))
}
