package syntax

// Linear reports whether m takes time linear in the length of the input, for
// each input. The optimizer can build a tree of matchers that takes time of
// the order of the length of the input to the power of the depth of the
// tree. [Node.Match] uses the [NFAMatcher] for such a tree (D20).
//
// The answer is conservative: a matcher that Linear does not know is not
// linear.
func Linear(m Matcher) bool {
	switch m := m.(type) {
	case TreeMatcher:
		return linearTree(m)
	case SizedTreeMatcher:
		return linearTree(m.TreeMatcher)
	case AnyOfMatcher:
		return allLinear(m.v)
	case IndexedAnyOfMatcher:
		return allLinear(m.AnyOfMatcher.v)
	case IndexedSizedAnyOfMatcher:
		return allLinear(m.AnyOfMatcher.v)
	case EveryOfMatcher:
		return allLinear(m.ms)
	case IndexedEveryOf:
		return allLinear(m.EveryOfMatcher.ms)
	case RowMatcher:
		return allBounded(m.ms)
	case AnyMatcher, ContainsMatcher, PrefixAnyMatcher, SuffixAnyMatcher:
		return true
	}
	return bounded(m)
}

// linearTree reports whether the tree takes time linear in the length of the
// input. In the modes treeFirst and treeLast, the tree finds each occurrence
// of the value, and matches one side once. In the mode treeEvery, it matches
// each side for each occurrence, so each side must take bounded time. A left
// side that matches nothing stops the search after the first occurrence, so
// the value is indexed once.
func linearTree(t TreeMatcher) bool {
	switch t.mode {
	case treeFirst:
		return fixed(t.value) && Linear(t.right)
	case treeLast:
		return fixed(t.value) && Linear(t.left)
	}
	if _, ok := t.left.(NothingMatcher); ok {
		return indexLinear(t.value) && (side(t.right) || fixed(t.value) && Linear(t.right))
	}
	if _, ok := t.right.(NothingMatcher); ok && fixed(t.value) {
		// The left side is matched only for the occurrence that ends the
		// input. See [TreeMatcher.Match].
		return Linear(t.left)
	}
	return fixed(t.value) && side(t.left) && side(t.right)
}

// side reports whether a tree answers the side m in bounded time. See
// [treeSides].
func side(m Matcher) bool {
	_, ok := m.(AnyMatcher)
	return ok || bounded(m)
}

// indexLinear reports whether one call of Index of v takes time linear in
// the length of the input.
func indexLinear(v Matcher) bool {
	switch v := v.(type) {
	case PrefixMatcher, SuffixMatcher, PrefixSuffixMatcher, ContainsMatcher,
		AnyMatcher, SuperMatcher, PrefixAnyMatcher, MinMatcher, MaxMatcher:
		return true
	case IndexedAnyOfMatcher:
		return allIndexLinear(v.v)
	case IndexedSizedAnyOfMatcher:
		return allIndexLinear(v.v)
	}
	return fixed(v)
}

// allIndexLinear reports whether Index of each matcher of v takes linear
// time.
func allIndexLinear(v []MatchIndexer) bool {
	for _, m := range v {
		if !indexLinear(m) {
			return false
		}
	}
	return true
}

// fixed reports whether each occurrence of v that Index finds has one
// length, and Index takes time linear in the distance to the occurrence.
func fixed(v Matcher) bool {
	switch v := v.(type) {
	case TextMatcher, SingleMatcher, ListMatcher, RangeMatcher, ClassMatcher:
		return true
	case RowMatcher:
		return allBounded(v.ms)
	}
	return false
}

// bounded reports whether m takes time that does not grow with the length
// of the input.
func bounded(m Matcher) bool {
	switch m := m.(type) {
	case NothingMatcher, SuperMatcher, TextMatcher, PrefixMatcher, SuffixMatcher,
		PrefixSuffixMatcher, SingleMatcher, ListMatcher, RangeMatcher,
		ClassMatcher, MinMatcher, MaxMatcher:
		return true
	case RowMatcher:
		return allBounded(m.ms)
	case AnyOfMatcher:
		return allBounded(m.v)
	case IndexedAnyOfMatcher:
		return allBounded(m.AnyOfMatcher.v)
	case IndexedSizedAnyOfMatcher:
		return allBounded(m.AnyOfMatcher.v)
	}
	return false
}

// allLinear reports whether each matcher of v is linear.
func allLinear[T Matcher](v []T) bool {
	for _, m := range v {
		if !Linear(m) {
			return false
		}
	}
	return true
}

// allBounded reports whether each matcher of v takes bounded time.
func allBounded[T Matcher](v []T) bool {
	for _, m := range v {
		if !bounded(m) {
			return false
		}
	}
	return true
}
