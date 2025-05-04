package match

type NothingMatcher struct{}

func NewNothing() NothingMatcher {
	return NothingMatcher{}
}

func (NothingMatcher) Match(s string) bool {
	return len(s) == 0
}

func (NothingMatcher) Index(s string) (int, []int) {
	return 0, segments0
}

func (NothingMatcher) Len() int {
	return 0
}

func (NothingMatcher) Size() int {
	return 0
}

func (NothingMatcher) String() string {
	return "<nothing>"
}
