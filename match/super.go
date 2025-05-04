package match

import (
	"fmt"
)

type SuperMatcher struct{}

func NewSuper() SuperMatcher {
	return SuperMatcher{}
}

func (s SuperMatcher) Match(_ string) bool {
	return true
}

func (s SuperMatcher) Len() int {
	return 0
}

func (s SuperMatcher) Index(v string) (int, []int) {
	seg := acquireSegments(len(v) + 1)
	for i := range v {
		seg = append(seg, i)
	}
	seg = append(seg, len(v))
	return 0, seg
}

func (s SuperMatcher) String() string {
	return fmt.Sprintf("<super>")
}
