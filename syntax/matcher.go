package syntax

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"
)

type AnyMatcher struct {
	sep []rune
}

func NewAny(s []rune) AnyMatcher {
	return AnyMatcher{s}
}

func (m AnyMatcher) Match(s string) bool {
	return runesIndexAnyRune(s, m.sep) == -1
}

func (m AnyMatcher) Index(s string) (int, []int) {
	switch i := runesIndexAnyRune(s, m.sep); i {
	case -1:
	case 0:
		return 0, segments0
	default:
		s = s[:i]
	}
	v := acquireSegments(len(s))
	for i := range s {
		v = append(v, i)
	}
	return 0, append(v, len(s))
}

func (AnyMatcher) Len() int {
	return 0
}

func (m AnyMatcher) String() string {
	return fmt.Sprintf("<any:![%s]>", string(m.sep))
}

type AnyOfMatcher struct {
	v []Matcher
	n int
}

func NewAnyOf(v ...Matcher) Matcher {
	// determine minimum
	var minimum int
	for i, m := range v {
		n := m.Len()
		if i == 0 || n < minimum {
			minimum = n
		}
	}
	a := AnyOfMatcher{v, minimum}
	if mis, ok := MatchIndexers(v); ok {
		x := IndexedAnyOfMatcher{a, mis}
		if msz, ok := MatchIndexSizers(v); ok {
			sz := -1
			for _, m := range msz {
				n := m.Size()
				if sz == -1 {
					sz = n
				} else if sz != n {
					sz = -1
					break
				}
			}
			if sz != -1 {
				return IndexedSizedAnyOfMatcher{x, sz}
			}
		}
		return x
	}
	return a
}

func MustIndexedAnyOf(v ...Matcher) MatchIndexer {
	return NewAnyOf(v...).(MatchIndexer)
}

func MustIndexedSizedAnyOf(v ...Matcher) MatchIndexSizer {
	return NewAnyOf(v...).(MatchIndexSizer)
}

func (m AnyOfMatcher) Match(s string) (ok bool) {
	if debugEnabled {
		done := debugMatching("any_of", s)
		defer func() {
			done(ok)
		}()
	}
	for _, matcher := range m.v {
		if matcher.Match(s) {
			return true
		}
	}
	return false
}

func (m AnyOfMatcher) Len() (n int) {
	return m.n
}

func (m AnyOfMatcher) Content(f func(Matcher)) {
	for _, matcher := range m.v {
		f(matcher)
	}
}

// String satisfies the [fmt.Stringer] interface.
func (m AnyOfMatcher) String() string {
	return fmt.Sprintf("<any_of:[%s]>", Matchers(m.v))
}

type IndexedAnyOfMatcher struct {
	AnyOfMatcher
	v []MatchIndexer
}

func (m IndexedAnyOfMatcher) Index(s string) (index int, segments []int) {
	if debugEnabled {
		done := debugIndexing("any_of", s)
		defer func() {
			done(index, segments)
		}()
	}
	index = -1
	segments = acquireSegments(len(s))
	for _, matcher := range m.v {
		if debugEnabled {
			debugLogf("indexing: any_of: trying %s", matcher)
		}
		i, seg := matcher.Index(s)
		if i == -1 {
			continue
		}
		if index == -1 || i < index {
			index = i
			segments = append(segments[:0], seg...)
			continue
		}
		if i > index {
			continue
		}
		// here i == index
		segments = appendMerge(segments, seg)
	}
	if index == -1 {
		releaseSegments(segments)
		return -1, nil
	}
	return index, segments
}

// String satisfies the [fmt.Stringer] interface.
func (m IndexedAnyOfMatcher) String() string {
	return fmt.Sprintf("<indexed_any_of:[%s]>", m.v)
}

type IndexedSizedAnyOfMatcher struct {
	IndexedAnyOfMatcher
	runes int
}

func (m IndexedSizedAnyOfMatcher) Size() int {
	return m.runes
}

type ContainsMatcher struct {
	s   string
	not bool
}

func NewContains(needle string) ContainsMatcher {
	return ContainsMatcher{needle, false}
}

func NewNotContains(needle string) ContainsMatcher {
	return ContainsMatcher{needle, true}
}

func (m ContainsMatcher) Match(s string) bool {
	return strings.Contains(s, m.s) != m.not
}

func (m ContainsMatcher) Index(s string) (int, []int) {
	var offset int
	idx := strings.Index(s, m.s)
	if !m.not {
		if idx == -1 {
			return -1, nil
		}
		offset = idx + len(m.s)
		if len(s) <= offset {
			return 0, []int{offset}
		}
		s = s[offset:]
	} else if idx != -1 {
		s = s[:idx]
	}
	segments := acquireSegments(len(s) + 1)
	for i := range s {
		segments = append(segments, offset+i)
	}
	return 0, append(segments, offset+len(s))
}

func (m ContainsMatcher) Len() int {
	return 0
}

// String satisfies the [fmt.Stringer] interface.
func (m ContainsMatcher) String() string {
	var not string
	if m.not {
		not = "!"
	}
	return fmt.Sprintf("<contains:%s[%s]>", not, m.s)
}

type EveryOfMatcher struct {
	ms  []Matcher
	min int
}

func NewEveryOf(ms []Matcher) Matcher {
	maximum := 0
	for i, m := range ms {
		n := m.Len()
		if i == 0 || n > maximum {
			maximum = n
		}
	}
	e := EveryOfMatcher{
		ms:  ms,
		min: maximum,
	}
	if mis, ok := MatchIndexers(ms); ok {
		return IndexedEveryOf{e, mis}
	}
	return e
}

func (m EveryOfMatcher) Len() (n int) {
	return m.min
}

func (m EveryOfMatcher) Match(s string) bool {
	for _, m := range m.ms {
		if !m.Match(s) {
			return false
		}
	}
	return true
}

func (m EveryOfMatcher) Content(f func(Matcher)) {
	for _, m := range m.ms {
		f(m)
	}
}

// String satisfies the [fmt.Stringer] interface.
func (m EveryOfMatcher) String() string {
	return fmt.Sprintf("<every_of:[%s]>", m.ms)
}

type IndexedEveryOf struct {
	EveryOfMatcher
	ms []MatchIndexer
}

func (m IndexedEveryOf) Index(s string) (int, []int) {
	// Each matcher reports the first index where it can match. The index of
	// all of them is the first index where each can match, with a length
	// that each can match.
	for i := 0; i <= len(s); {
		var (
			current []int
			skip    int
		)
		for k, mi := range m.ms {
			idx, seg := mi.Index(s[i:])
			if idx == -1 {
				return -1, nil
			}
			if idx > 0 {
				skip = max(skip, idx)
				continue
			}
			if skip != 0 {
				continue
			}
			if k == 0 {
				current = append(current, seg...)
				continue
			}
			current = slices.DeleteFunc(current, func(n int) bool {
				return !slices.Contains(seg, n)
			})
		}
		switch {
		case skip != 0:
			i += skip
		case len(current) != 0:
			return i, current
		case i == len(s):
			return -1, nil
		default:
			_, w := utf8.DecodeRuneInString(s[i:])
			i += w
		}
	}
	return -1, nil
}

// String satisfies the [fmt.Stringer] interface.
func (m IndexedEveryOf) String() string {
	return fmt.Sprintf("<indexed_every_of:[%s]>", m.ms)
}

type ListMatcher struct {
	rs  []rune
	not bool
}

func NewList(rs []rune, not bool) ListMatcher {
	return ListMatcher{rs, not}
}

func (m ListMatcher) Match(s string) bool {
	r, w := utf8.DecodeRuneInString(s)
	if w == 0 || len(s) > w {
		// Invalid rune.
		return false
	}
	inList := runesIndexRune(m.rs, r) != -1
	return inList == !m.not
}

func (ListMatcher) Len() int {
	return 1
}

func (ListMatcher) Size() int {
	return 1
}

func (m ListMatcher) Index(s string) (int, []int) {
	for i, r := range s {
		if m.not == (runesIndexRune(m.rs, r) == -1) {
			return i, segmentsByRuneLength[runeWidth(s, i)]
		}
	}
	return -1, nil
}

// String satisfies the [fmt.Stringer] interface.
func (m ListMatcher) String() string {
	var not string
	if m.not {
		not = "!"
	}
	return fmt.Sprintf("<list:%s[%s]>", not, string(m.rs))
}

type MaxMatcher struct {
	n int
}

func NewMax(n int) MaxMatcher {
	return MaxMatcher{n}
}

func (m MaxMatcher) Match(s string) bool {
	var n int
	for range s {
		n += 1
		if n > m.n {
			return false
		}
	}
	return true
}

func (m MaxMatcher) Index(s string) (int, []int) {
	segments := acquireSegments(m.n + 1)
	segments = append(segments, 0)
	var count int
	for i := range s {
		count++
		if count > m.n {
			break
		}
		segments = append(segments, i+runeWidth(s, i))
	}
	return 0, segments
}

func (m MaxMatcher) Len() int {
	return 0
}

// String satisfies the [fmt.Stringer] interface.
func (m MaxMatcher) String() string {
	return fmt.Sprintf("<max:%d>", m.n)
}

type MinMatcher struct {
	n int
}

func NewMin(n int) MinMatcher {
	return MinMatcher{n}
}

func (m MinMatcher) Match(s string) bool {
	var n int
	for range s {
		n += 1
		if n >= m.n {
			return true
		}
	}
	return false
}

func (m MinMatcher) Index(s string) (int, []int) {
	var count int
	c := len(s) - m.n + 1
	if c <= 0 {
		return -1, nil
	}
	segments := acquireSegments(c)
	for i := range s {
		count++
		if count >= m.n {
			segments = append(segments, i+runeWidth(s, i))
		}
	}
	if len(segments) == 0 {
		return -1, nil
	}
	return 0, segments
}

func (m MinMatcher) Len() int {
	return m.n
}

// String satisfies the [fmt.Stringer] interface.
func (m MinMatcher) String() string {
	return fmt.Sprintf("<min:%d>", m.n)
}

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

// String satisfies the [fmt.Stringer] interface.
func (NothingMatcher) String() string {
	return "<nothing>"
}

type PrefixAnyMatcher struct {
	s   string
	sep []rune
	n   int
}

func NewPrefixAny(s string, sep []rune) PrefixAnyMatcher {
	return PrefixAnyMatcher{s, sep, utf8.RuneCountInString(s)}
}

func (m PrefixAnyMatcher) Index(s string) (int, []int) {
	idx := strings.Index(s, m.s)
	if idx == -1 {
		return -1, nil
	}
	n := len(m.s)
	sub := s[idx+n:]
	i := runesIndexAnyRune(sub, m.sep)
	if i > -1 {
		sub = sub[:i]
	}
	seg := acquireSegments(len(sub) + 1)
	seg = append(seg, n)
	for i := range sub {
		seg = append(seg, n+i+runeWidth(sub, i))
	}
	return idx, seg
}

func (m PrefixAnyMatcher) Len() int {
	return m.n
}

func (m PrefixAnyMatcher) Match(s string) bool {
	if !strings.HasPrefix(s, m.s) {
		return false
	}
	return runesIndexAnyRune(s[len(m.s):], m.sep) == -1
}

// String satisfies the [fmt.Stringer] interface.
func (m PrefixAnyMatcher) String() string {
	return fmt.Sprintf("<prefix_any:%s![%s]>", m.s, string(m.sep))
}

type PrefixMatcher struct {
	s string
	n int
}

func NewPrefix(p string) PrefixMatcher {
	return PrefixMatcher{
		s: p,
		n: utf8.RuneCountInString(p),
	}
}

func (m PrefixMatcher) Index(s string) (int, []int) {
	idx := strings.Index(s, m.s)
	if idx == -1 {
		return -1, nil
	}
	length := len(m.s)
	var sub string
	if len(s) > idx+length {
		sub = s[idx+length:]
	} else {
		sub = ""
	}
	segments := acquireSegments(len(sub) + 1)
	segments = append(segments, length)
	for i := range sub {
		segments = append(segments, length+i+runeWidth(sub, i))
	}
	return idx, segments
}

func (m PrefixMatcher) Len() int {
	return m.n
}

func (m PrefixMatcher) Match(s string) bool {
	return strings.HasPrefix(s, m.s)
}

// String satisfies the [fmt.Stringer] interface.
func (m PrefixMatcher) String() string {
	return fmt.Sprintf("<prefix:%s>", m.s)
}

type PrefixSuffixMatcher struct {
	p, s string
	n    int
}

func NewPrefixSuffix(prefix, suffix string) PrefixSuffixMatcher {
	pn := utf8.RuneCountInString(prefix)
	sn := utf8.RuneCountInString(suffix)
	return PrefixSuffixMatcher{prefix, suffix, pn + sn}
}

func (m PrefixSuffixMatcher) Index(s string) (int, []int) {
	prefixIdx := strings.Index(s, m.p)
	if prefixIdx == -1 {
		return -1, nil
	}
	suffixLen := len(m.s)
	if suffixLen <= 0 {
		return prefixIdx, []int{len(s) - prefixIdx}
	}
	// Each occurrence of the suffix after the prefix ends a match.
	// Occurrences can overlap each other, and must not overlap the prefix.
	sub := s[prefixIdx:]
	i := strings.Index(sub[len(m.p):], m.s)
	if i == -1 {
		return -1, nil
	}
	return prefixIdx, suffixEnds(sub, m.s, len(m.p)+i)
}

func (m PrefixSuffixMatcher) Match(s string) bool {
	return len(s) >= len(m.p)+len(m.s) && strings.HasPrefix(s, m.p) && strings.HasSuffix(s, m.s)
}

func (m PrefixSuffixMatcher) Len() int {
	return m.n
}

// String satisfies the [fmt.Stringer] interface.
func (m PrefixSuffixMatcher) String() string {
	return fmt.Sprintf("<prefix_suffix:[%s,%s]>", m.p, m.s)
}

type RangeMatcher struct {
	Lo, Hi rune
	Not    bool
}

func NewRange(lo, hi rune, not bool) RangeMatcher {
	return RangeMatcher{lo, hi, not}
}

func (RangeMatcher) Len() int {
	return 1
}

func (RangeMatcher) Size() int {
	return 1
}

func (m RangeMatcher) Match(s string) (ok bool) {
	if debugEnabled {
		done := debugMatching("range", s)
		defer func() { done(ok) }()
	}
	r, w := utf8.DecodeRuneInString(s)
	if w == 0 || len(s) > w {
		return false
	}
	inRange := r >= m.Lo && r <= m.Hi
	return inRange == !m.Not
}

func (m RangeMatcher) Index(s string) (index int, segments []int) {
	if debugEnabled {
		done := debugIndexing("range", s)
		defer func() { done(index, segments) }()
	}
	for i, r := range s {
		if m.Not != (r >= m.Lo && r <= m.Hi) {
			return i, segmentsByRuneLength[runeWidth(s, i)]
		}
	}
	return -1, nil
}

// String satisfies the [fmt.Stringer] interface.
func (m RangeMatcher) String() string {
	var not string
	if m.Not {
		not = "!"
	}
	return fmt.Sprintf("<range:%s[%s,%s]>", not, string(m.Lo), string(m.Hi))
}

type RowMatcher struct {
	ms  []MatchIndexSizer
	n   int
	seg []int
}

func NewRow(ms []MatchIndexSizer) RowMatcher {
	var r int
	for _, m := range ms {
		r += m.Size()
	}
	return RowMatcher{
		ms:  ms,
		n:   r,
		seg: []int{r},
	}
}

func (m RowMatcher) Match(s string) (ok bool) {
	if debugEnabled {
		done := debugMatching("row", s)
		defer func() { done(ok) }()
	}
	if !runesExactlyRunesCount(s, m.n) {
		return false
	}
	return m.matchAll(s)
}

func (m RowMatcher) Len() int {
	return m.n
}

func (m RowMatcher) Size() int {
	return m.n
}

func (m RowMatcher) Index(s string) (index int, segments []int) {
	if debugEnabled {
		done := debugIndexing("row", s)
		debugLogf("row: %d vs %d", len(s), m.n)
		defer func() { done(index, segments) }()
	}
	for j := 0; j <= len(s)-m.n; { // NOTE: using len() here to avoid counting runes.
		i, _ := m.ms[0].Index(s[j:])
		if i == -1 {
			return -1, nil
		}
		i += j
		if runesHasRunes(s[i:], m.n) && m.matchAll(s[i:]) {
			// A segment is a length in bytes, and the row is m.n runes.
			if n := len(runesHead(s[i:], m.n)); n != m.n {
				return i, []int{n}
			}
			return i, m.seg
		}
		_, x := utf8.DecodeRuneInString(s[i:])
		j = i + max(x, 1)
	}
	return -1, nil
}

func (m RowMatcher) Content(f func(Matcher)) {
	for _, matcher := range m.ms {
		f(matcher)
	}
}

// String satisfies the [fmt.Stringer] interface.
func (m RowMatcher) String() string {
	return fmt.Sprintf("<row_%d:%s>", m.n, m.ms)
}

func (m RowMatcher) matchAll(s string) bool {
	var i int
	for _, m := range m.ms {
		n := m.Size()
		sub := runesHead(s[i:], n)
		if !m.Match(sub) {
			return false
		}
		i += len(sub)
	}
	return true
}

// single represents ?
type SingleMatcher struct {
	sep []rune
}

func NewSingle(s []rune) SingleMatcher {
	return SingleMatcher{s}
}

func (m SingleMatcher) Match(v string) bool {
	r, w := utf8.DecodeRuneInString(v)
	if w == 0 || len(v) > w {
		return false
	}
	return runesIndexRune(m.sep, r) == -1
}

func (SingleMatcher) Len() int {
	return 1
}

func (SingleMatcher) Size() int {
	return 1
}

func (m SingleMatcher) Index(v string) (int, []int) {
	for i, r := range v {
		if runesIndexRune(m.sep, r) == -1 {
			return i, segmentsByRuneLength[runeWidth(v, i)]
		}
	}
	return -1, nil
}

// String satisfies the [fmt.Stringer] interface.
func (m SingleMatcher) String() string {
	if len(m.sep) == 0 {
		return "<single>"
	}
	return fmt.Sprintf("<single:![%s]>", string(m.sep))
}

type SuffixAnyMatcher struct {
	s   string
	sep []rune
	n   int
}

func NewSuffixAny(s string, sep []rune) SuffixAnyMatcher {
	return SuffixAnyMatcher{s, sep, utf8.RuneCountInString(s)}
}

func (m SuffixAnyMatcher) Index(v string) (int, []int) {
	idx := strings.Index(v, m.s)
	if idx == -1 {
		return -1, nil
	}
	i := runesLastIndexAnyRune(v[:idx], m.sep) + 1
	// Each occurrence of the suffix ends a match, until a separator.
	var seg []int
	for _, p := range suffixEnds(v, m.s, idx) {
		if runesIndexAnyRune(v[i:p-len(m.s)], m.sep) != -1 {
			break
		}
		seg = append(seg, p-i)
	}
	return i, seg
}

// suffixEnds returns the end of each occurrence of suffix in v, starting with
// the occurrence at idx. Occurrences can overlap.
func suffixEnds(v, suffix string, idx int) []int {
	var ends []int
	for idx != -1 {
		ends = append(ends, idx+len(suffix))
		_, w := utf8.DecodeRuneInString(v[idx:])
		if w == 0 {
			break
		}
		next := strings.Index(v[idx+w:], suffix)
		if next == -1 {
			break
		}
		idx += w + next
	}
	return ends
}

func (m SuffixAnyMatcher) Len() int {
	return m.n
}

func (m SuffixAnyMatcher) Match(v string) bool {
	if !strings.HasSuffix(v, m.s) {
		return false
	}
	return runesIndexAnyRune(v[:len(v)-len(m.s)], m.sep) == -1
}

// String satisfies the [fmt.Stringer] interface.
func (m SuffixAnyMatcher) String() string {
	return fmt.Sprintf("<suffix_any:![%s]%s>", string(m.sep), m.s)
}

type SuffixMatcher struct {
	s string
	n int
}

func NewSuffix(s string) SuffixMatcher {
	return SuffixMatcher{s, utf8.RuneCountInString(s)}
}

func (m SuffixMatcher) Len() int {
	return m.n
}

func (m SuffixMatcher) Match(v string) bool {
	return strings.HasSuffix(v, m.s)
}

func (m SuffixMatcher) Index(v string) (int, []int) {
	if i := strings.Index(v, m.s); i != -1 {
		return 0, suffixEnds(v, m.s, i)
	}
	return -1, nil
}

// String satisfies the [fmt.Stringer] interface.
func (m SuffixMatcher) String() string {
	return fmt.Sprintf("<suffix:%s>", m.s)
}

type SuperMatcher struct{}

func NewSuper() SuperMatcher {
	return SuperMatcher{}
}

func (m SuperMatcher) Match(_ string) bool {
	return true
}

func (m SuperMatcher) Len() int {
	return 0
}

func (m SuperMatcher) Index(v string) (int, []int) {
	seg := acquireSegments(len(v) + 1)
	for i := range v {
		seg = append(seg, i)
	}
	seg = append(seg, len(v))
	return 0, seg
}

// String satisfies the [fmt.Stringer] interface.
func (m SuperMatcher) String() string {
	return fmt.Sprintf("<super>")
}

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

func (m TextMatcher) Match(s string) bool {
	return m.s == s
}

func (m TextMatcher) Index(s string) (int, []int) {
	i := strings.Index(s, m.s)
	if i == -1 {
		return -1, nil
	}
	return i, m.seg
}

func (m TextMatcher) Len() int {
	return m.runes
}

func (m TextMatcher) BytesCount() int {
	return m.bytes
}

func (m TextMatcher) Size() int {
	return m.runes
}

// String satisfies the [fmt.Stringer] interface.
func (m TextMatcher) String() string {
	return fmt.Sprintf("<text:`%v`>", m.s)
}

type TreeMatcher struct {
	value  MatchIndexer
	left   Matcher
	right  Matcher
	minLen int
	runes  int
	vrunes int
	lrunes int
	rrunes int
	// mode tells Match which occurrences of the value it must try.
	mode treeMode
}

// treeMode tells [TreeMatcher.Match] which occurrences of the value it must
// try. Without a mode, a tree of n stars takes time of the order of the
// length of the input to the power n.
type treeMode uint8

const (
	// treeEvery tries every occurrence of the value.
	treeEvery treeMode = iota
	// treeFirst tries only the occurrence that ends first. The left side
	// matches any prefix, and the right side matches each string that ends
	// with a string that it matches. A longer rest can only help.
	treeFirst
	// treeLast tries only the occurrence that starts last. It is treeFirst
	// with the sides swapped.
	treeLast
)

func NewTree(v MatchIndexer, l, r Matcher) Matcher {
	tree := TreeMatcher{
		value: v,
		left:  l,
		right: r,
	}
	tree.minLen = v.Len()
	if l != nil {
		tree.minLen += l.Len()
	}
	if r != nil {
		tree.minLen += r.Len()
	}
	var (
		ls, lsz = l.(Sizer)
		rs, rsz = r.(Sizer)
		vs, vsz = v.(Sizer)
	)
	if lsz {
		tree.lrunes = ls.Size()
	}
	if rsz {
		tree.rrunes = rs.Size()
	}
	if vsz {
		tree.vrunes = vs.Size()
	}
	// li, lix := l.(MatchIndexer)
	// ri, rix := r.(MatchIndexer)
	tree.mode = newTreeMode(v, l, r)
	if vsz && lsz && rsz {
		tree.runes = tree.vrunes + tree.lrunes + tree.rrunes
		return SizedTreeMatcher{tree}
	}
	return tree
}

// newTreeMode returns the mode of a tree of v, l and r.
func newTreeMode(v MatchIndexer, l, r Matcher) treeMode {
	if star, ok := starSeparators(l); ok && openLeft(r, star) && textWithout(v, star) {
		return treeFirst
	}
	if star, ok := starSeparators(r); ok && openRight(l, star) && textWithout(v, star) {
		return treeLast
	}
	return treeEvery
}

// starSeparators reports whether m is a star, and returns the separators that
// the star cannot match. A star `**` has none.
func starSeparators(m Matcher) ([]rune, bool) {
	switch m := m.(type) {
	case SuperMatcher:
		return nil, true
	case AnyMatcher:
		return m.sep, true
	}
	return nil, false
}

// textWithout reports whether no match of v can hold a rune of sep. It is
// true for each v when sep is empty.
func textWithout(v MatchIndexer, sep []rune) bool {
	if len(sep) == 0 {
		return true
	}
	t, ok := v.(TextMatcher)
	return ok && !strings.ContainsFunc(t.s, func(r rune) bool { return slices.Contains(sep, r) })
}

// subset reports whether each rune of a is in b.
func subset(a, b []rune) bool {
	for _, r := range a {
		if !slices.Contains(b, r) {
			return false
		}
	}
	return true
}

// openLeft reports whether m matches x+s whenever it matches s, for each x
// that holds no rune of sep. An empty sep means each x.
func openLeft(m Matcher, sep []rune) bool {
	switch m := m.(type) {
	case SuperMatcher, SuffixMatcher:
		return true
	case ContainsMatcher:
		return !m.not
	case AnyMatcher:
		return len(sep) != 0 && subset(m.sep, sep)
	case SuffixAnyMatcher:
		return len(sep) != 0 && subset(m.sep, sep)
	case TreeMatcher:
		return m.lrunes == 0 && openLeft(m.left, sep)
	case SizedTreeMatcher:
		return false
	case AnyOfMatcher:
		return allOpen(m.v, sep, openLeft)
	case IndexedAnyOfMatcher:
		return allOpen(m.AnyOfMatcher.v, sep, openLeft)
	case IndexedSizedAnyOfMatcher:
		return allOpen(m.AnyOfMatcher.v, sep, openLeft)
	}
	return false
}

// openRight reports whether m matches s+x whenever it matches s, for each x
// that holds no rune of sep. An empty sep means each x.
func openRight(m Matcher, sep []rune) bool {
	switch m := m.(type) {
	case SuperMatcher, PrefixMatcher:
		return true
	case ContainsMatcher:
		return !m.not
	case AnyMatcher:
		return len(sep) != 0 && subset(m.sep, sep)
	case PrefixAnyMatcher:
		return len(sep) != 0 && subset(m.sep, sep)
	case TreeMatcher:
		return m.rrunes == 0 && openRight(m.right, sep)
	case SizedTreeMatcher:
		return false
	case AnyOfMatcher:
		return allOpen(m.v, sep, openRight)
	case IndexedAnyOfMatcher:
		return allOpen(m.AnyOfMatcher.v, sep, openRight)
	case IndexedSizedAnyOfMatcher:
		return allOpen(m.AnyOfMatcher.v, sep, openRight)
	}
	return false
}

// allOpen reports whether open is true for each matcher of v.
func allOpen(v []Matcher, sep []rune, open func(Matcher, []rune) bool) bool {
	for _, m := range v {
		if !open(m, sep) {
			return false
		}
	}
	return true
}

func (m TreeMatcher) Len() int {
	return m.minLen
}

func (m TreeMatcher) Content(f func(Matcher)) {
	if m.left != nil {
		f(m.left)
	}
	f(m.value)
	if m.right != nil {
		f(m.right)
	}
}

func (m TreeMatcher) Match(s string) (ok bool) {
	if debugEnabled {
		done := debugMatching("tree", s)
		defer func() { done(ok) }()
	}
	switch m.mode {
	case treeFirst:
		return m.matchFirst(s)
	case treeLast:
		return m.matchLast(s)
	}
	n := len(s)
	offset, limit := m.offsetLimit(s)
	sides := newTreeSides(m.left, m.right, s)
	lazy := !side(m.left) && side(m.right)
	for len(s)-offset-limit >= m.vrunes {
		if debugEnabled {
			debugLogf(
				"value %s indexing: %q (offset=%d; limit=%d)",
				m.value, s[offset:n-limit], offset, limit,
			)
		}
		index, segments := m.value.Index(s[offset : n-limit])
		if debugEnabled {
			debugLogf(
				"value %s index: %d; %v",
				m.value, index, segments,
			)
		}
		if index == -1 {
			releaseSegments(segments)
			return false
		}
		index += offset
		// A left side that takes time is matched only after a right side
		// matched, so that a right side such as nothing saves the time.
		var left, stop bool
		if !lazy {
			left, stop = sides.left(m.left, s, index)
		}
		if stop {
			// No longer left side can match either.
			releaseSegments(segments)
			return false
		}
		if left || lazy {
			for _, seg := range segments {
				if !sides.right(m.right, s, index+seg) {
					continue
				}
				if lazy {
					if left, _ = sides.left(m.left, s, index); !left {
						break
					}
				}
				releaseSegments(segments)
				return true
			}
		}
		releaseSegments(segments)
		_, x := utf8.DecodeRuneInString(s[index:])
		if x == 0 {
			// No progress.
			break
		}
		offset = index + x
	}
	return false
}

// treeSides answers the matches of the sides of a tree. It answers a side
// that is a star or nothing from the positions of the separators, which it
// finds once for each input. Without them, each answer takes time linear in
// the length of the input.
type treeSides struct {
	firstSep, lastSep int
}

// newTreeSides returns the sides of a tree of l and r for the input s.
func newTreeSides(l, r Matcher, s string) treeSides {
	t := treeSides{firstSep: len(s), lastSep: -1}
	if a, ok := l.(AnyMatcher); ok {
		if i := runesIndexAnyRune(s, a.sep); i != -1 {
			t.firstSep = i
		}
	}
	if a, ok := r.(AnyMatcher); ok {
		t.lastSep = runesLastIndexAnyRune(s, a.sep)
	}
	return t
}

// left reports whether l matches s[:i]. It also reports whether l cannot
// match s[:j] for each j >= i.
func (t treeSides) left(l Matcher, s string, i int) (ok, stop bool) {
	switch l.(type) {
	case AnyMatcher:
		ok = i <= t.firstSep
		return ok, !ok
	case NothingMatcher:
		return i == 0, i != 0
	}
	if debugEnabled {
		debugLogf("matching left: %q", s[:i])
	}
	return l.Match(s[:i]), false
}

// right reports whether r matches s[i:].
func (t treeSides) right(r Matcher, s string, i int) bool {
	switch r.(type) {
	case AnyMatcher:
		return i > t.lastSep
	case NothingMatcher:
		return i == len(s)
	}
	if debugEnabled {
		debugLogf("matching right: %q", s[i:])
	}
	return r.Match(s[i:])
}

// matchFirst answers Match for the mode treeFirst. It finds the occurrence of
// the value that ends first, and matches the right side once.
func (m TreeMatcher) matchFirst(s string) bool {
	end := -1
	sides := newTreeSides(m.left, m.right, s)
	for offset := 0; offset <= len(s); {
		index, segments := m.value.Index(s[offset:])
		if index == -1 {
			break
		}
		index += offset
		if end != -1 && index >= end {
			releaseSegments(segments)
			break
		}
		if left, stop := sides.left(m.left, s, index); stop {
			releaseSegments(segments)
			break
		} else if left {
			for _, seg := range segments {
				if end == -1 || index+seg < end {
					end = index + seg
				}
			}
		}
		releaseSegments(segments)
		_, w := utf8.DecodeRuneInString(s[index:])
		if w == 0 {
			break
		}
		offset = index + w
	}
	return end != -1 && m.right.Match(s[end:])
}

// matchLast answers Match for the mode treeLast. It finds the occurrence of
// the value that starts last, and matches the left side once.
func (m TreeMatcher) matchLast(s string) bool {
	start := -1
	sides := newTreeSides(m.left, m.right, s)
	for offset := 0; offset <= len(s); {
		index, segments := m.value.Index(s[offset:])
		if index == -1 {
			break
		}
		index += offset
		for _, seg := range segments {
			if sides.right(m.right, s, index+seg) {
				start = index
				break
			}
		}
		releaseSegments(segments)
		_, w := utf8.DecodeRuneInString(s[index:])
		if w == 0 {
			break
		}
		offset = index + w
	}
	return start != -1 && m.left.Match(s[:start])
}

// Returns substring and offset/limit pair in bytes.
func (m TreeMatcher) offsetLimit(s string) (offset, limit int) {
	n := utf8.RuneCountInString(s)
	if m.runes > n {
		return 0, 0
	}
	if n := m.lrunes; n > 0 {
		offset = len(runesHead(s, n))
	}
	if n := m.rrunes; n > 0 {
		limit = len(runesTail(s, n))
	}
	return
}

// String satisfies the [fmt.Stringer] interface.
func (m TreeMatcher) String() string {
	return fmt.Sprintf(
		"<btree:[%v<-%s->%v]>",
		m.left, m.value, m.right,
	)
}

type SizedTreeMatcher struct {
	TreeMatcher
}

func (m SizedTreeMatcher) Size() int {
	return m.TreeMatcher.runes
}

type IndexedTreeMatcher struct {
	value MatchIndexer
	left  MatchIndexer
	right MatchIndexer
}

type SomePool interface {
	Get() []int
	Put([]int)
}

var segmentsPools [1024]sync.Pool

func toPowerOfTwo(v int) int {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

const (
	cacheFrom             = 16
	cacheToAndHigher      = 1024
	cacheFromIndex        = 15
	cacheToAndHigherIndex = 1023
)

var (
	segments0 = []int{0}
	segments1 = []int{1}
	segments2 = []int{2}
	segments3 = []int{3}
	segments4 = []int{4}
)

var segmentsByRuneLength [5][]int = [5][]int{
	0: segments0,
	1: segments1,
	2: segments2,
	3: segments3,
	4: segments4,
}

func init() {
	for i := cacheToAndHigher; i >= cacheFrom; i >>= 1 {
		func(i int) {
			segmentsPools[i-1] = sync.Pool{New: func() any {
				return make([]int, 0, i)
			}}
		}(i)
	}
}

func getTableIndex(c int) int {
	p := toPowerOfTwo(c)
	switch {
	case p >= cacheToAndHigher:
		return cacheToAndHigherIndex
	case p <= cacheFrom:
		return cacheFromIndex
	default:
		return p - 1
	}
}

func acquireSegments(c int) []int {
	// make []int with less capacity than cacheFrom
	// is faster than acquiring it from pool
	if c < cacheFrom {
		return make([]int, 0, c)
	}
	return segmentsPools[getTableIndex(c)].Get().([]int)[:0]
}

func releaseSegments(s []int) {
	c := cap(s)
	// make []int with less capacity than cacheFrom
	// is faster than acquiring it from pool
	if c < cacheFrom {
		return
	}
	segmentsPools[getTableIndex(c)].Put(s)
}
