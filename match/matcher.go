package match

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/kenshaw/glob/debug"
	"github.com/kenshaw/glob/runes"
)

type AnyMatcher struct {
	sep []rune
}

func NewAny(s []rune) AnyMatcher {
	return AnyMatcher{s}
}

func (m AnyMatcher) Match(s string) bool {
	return runes.IndexAnyRune(s, m.sep) == -1
}

func (m AnyMatcher) Index(s string) (int, []int) {
	switch i := runes.IndexAnyRune(s, m.sep); i {
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
	if debug.Enabled {
		done := debug.Matching("any_of", s)
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
	if debug.Enabled {
		done := debug.Indexing("any_of", s)
		defer func() {
			done(index, segments)
		}()
	}
	index = -1
	segments = acquireSegments(len(s))
	for _, matcher := range m.v {
		if debug.Enabled {
			debug.Logf("indexing: any_of: trying %s", matcher)
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
	var index int
	var offset int
	// make `in` with cap as len(s),
	// cause it is the maximum size of output segments values
	next := acquireSegments(len(s))
	current := acquireSegments(len(s))
	sub := s
	for i, m := range m.ms {
		idx, seg := m.Index(sub)
		if idx == -1 {
			releaseSegments(next)
			releaseSegments(current)
			return -1, nil
		}
		if i == 0 {
			// we use copy here instead of `current = seg`
			// cause seg is a slice from reusable buffer `in`
			// and it could be overwritten in next iteration
			current = append(current, seg...)
		} else {
			// clear the next
			next = next[:0]
			delta := index - (idx + offset)
			for _, ex := range current {
				for _, n := range seg {
					if ex+delta == n {
						next = append(next, n)
					}
				}
			}
			if len(next) == 0 {
				releaseSegments(next)
				releaseSegments(current)
				return -1, nil
			}
			current = append(current[:0], next...)
		}
		index = idx + offset
		sub = s[index:]
		offset += idx
	}
	releaseSegments(next)
	return index, current
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
	if len(s) > w {
		// Invalid rune.
		return false
	}
	inList := runes.IndexRune(m.rs, r) != -1
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
		if m.not == (runes.IndexRune(m.rs, r) == -1) {
			return i, segmentsByRuneLength[utf8.RuneLen(r)]
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
	for i, r := range s {
		count++
		if count > m.n {
			break
		}
		segments = append(segments, i+utf8.RuneLen(r))
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
	for i, r := range s {
		count++
		if count >= m.n {
			segments = append(segments, i+utf8.RuneLen(r))
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
	i := runes.IndexAnyRune(sub, m.sep)
	if i > -1 {
		sub = sub[:i]
	}
	seg := acquireSegments(len(sub) + 1)
	seg = append(seg, n)
	for i, r := range sub {
		seg = append(seg, n+i+utf8.RuneLen(r))
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
	return runes.IndexAnyRune(s[len(m.s):], m.sep) == -1
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
	for i, r := range sub {
		segments = append(segments, length+i+utf8.RuneLen(r))
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
	if (len(s) - prefixIdx) <= 0 {
		return -1, nil
	}
	segments := acquireSegments(len(s) - prefixIdx)
	for sub := s[prefixIdx:]; ; {
		suffixIdx := strings.LastIndex(sub, m.s)
		if suffixIdx == -1 {
			break
		}
		segments = append(segments, suffixIdx+suffixLen)
		sub = sub[:suffixIdx]
	}
	if len(segments) == 0 {
		releaseSegments(segments)
		return -1, nil
	}
	slices.Reverse(segments)
	return prefixIdx, segments
}

func (m PrefixSuffixMatcher) Match(s string) bool {
	return strings.HasPrefix(s, m.p) && strings.HasSuffix(s, m.s)
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
	if debug.Enabled {
		done := debug.Matching("range", s)
		defer func() { done(ok) }()
	}
	r, w := utf8.DecodeRuneInString(s)
	if len(s) > w {
		return false
	}
	inRange := r >= m.Lo && r <= m.Hi
	return inRange == !m.Not
}

func (m RangeMatcher) Index(s string) (index int, segments []int) {
	if debug.Enabled {
		done := debug.Indexing("range", s)
		defer func() { done(index, segments) }()
	}
	for i, r := range s {
		if m.Not != (r >= m.Lo && r <= m.Hi) {
			return i, segmentsByRuneLength[utf8.RuneLen(r)]
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
	if debug.Enabled {
		done := debug.Matching("row", s)
		defer func() { done(ok) }()
	}
	if !runes.ExactlyRunesCount(s, m.n) {
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
	if debug.Enabled {
		done := debug.Indexing("row", s)
		debug.Logf("row: %d vs %d", len(s), m.n)
		defer func() { done(index, segments) }()
	}
	for j := 0; j <= len(s)-m.n; { // NOTE: using len() here to avoid counting runes.
		i, _ := m.ms[0].Index(s[j:])
		if i == -1 {
			return -1, nil
		}
		if m.matchAll(s[i:]) {
			return j + i, m.seg
		}
		_, x := utf8.DecodeRuneInString(s[i:])
		j += x
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
		sub := runes.Head(s[i:], n)
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
	if len(v) > w {
		return false
	}
	return runes.IndexRune(m.sep, r) == -1
}

func (SingleMatcher) Len() int {
	return 1
}

func (SingleMatcher) Size() int {
	return 1
}

func (m SingleMatcher) Index(v string) (int, []int) {
	for i, r := range v {
		if runes.IndexRune(m.sep, r) == -1 {
			return i, segmentsByRuneLength[utf8.RuneLen(r)]
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
	i := runes.LastIndexAnyRune(v[:idx], m.sep) + 1
	return i, []int{idx + len(m.s) - i}
}

func (m SuffixAnyMatcher) Len() int {
	return m.n
}

func (m SuffixAnyMatcher) Match(v string) bool {
	if !strings.HasSuffix(v, m.s) {
		return false
	}
	return runes.IndexAnyRune(v[:len(v)-len(m.s)], m.sep) == -1
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
		return 0, []int{i + len(m.s)}
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
}

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
	if vsz && lsz && rsz {
		tree.runes = tree.vrunes + tree.lrunes + tree.rrunes
		return SizedTreeMatcher{tree}
	}
	return tree
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
	if debug.Enabled {
		done := debug.Matching("tree", s)
		defer func() { done(ok) }()
	}
	n := len(s)
	offset, limit := m.offsetLimit(s)
	for len(s)-offset-limit >= m.vrunes {
		if debug.Enabled {
			debug.Logf(
				"value %s indexing: %q (offset=%d; limit=%d)",
				m.value, s[offset:n-limit], offset, limit,
			)
		}
		index, segments := m.value.Index(s[offset : n-limit])
		if debug.Enabled {
			debug.Logf(
				"value %s index: %d; %v",
				m.value, index, segments,
			)
		}
		if index == -1 {
			releaseSegments(segments)
			return false
		}
		if debug.Enabled {
			debug.Logf("matching left: %q", s[:offset+index])
		}
		left := m.left.Match(s[:offset+index])
		if debug.Enabled {
			debug.Logf("matching left: -> %t", left)
		}
		if left {
			for _, seg := range segments {
				if debug.Enabled {
					debug.Logf("matching right: %q", s[offset+index+seg:])
				}
				right := m.right.Match(s[offset+index+seg:])
				if debug.Enabled {
					debug.Logf("matching right: -> %t", right)
				}
				if right {
					releaseSegments(segments)
					return true
				}
			}
		}
		releaseSegments(segments)
		_, x := utf8.DecodeRuneInString(s[offset+index:])
		if x == 0 {
			// No progress.
			break
		}
		offset = offset + index + x
	}
	return false
}

// Returns substring and offset/limit pair in bytes.
func (m TreeMatcher) offsetLimit(s string) (offset, limit int) {
	n := utf8.RuneCountInString(s)
	if m.runes > n {
		return 0, 0
	}
	if n := m.lrunes; n > 0 {
		offset = len(runes.Head(s, n))
	}
	if n := m.rrunes; n > 0 {
		limit = len(runes.Tail(s, n))
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
