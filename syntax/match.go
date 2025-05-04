package syntax

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
)

// todo common table of rune's length

type Matcher interface {
	Match(string) bool
	Len() int
}

type Matchers []Matcher

// String satisfies the [fmt.Stringer] interface.
func (matchers Matchers) String() string {
	var s []string
	for _, matcher := range matchers {
		s = append(s, fmt.Sprint(matcher))
	}
	return fmt.Sprintf("%s", strings.Join(s, ","))
}

type Indexer interface {
	Index(string) (int, []int)
}

type Sizer interface {
	Size() int
}

type MatchSizer interface {
	Matcher
	Sizer
}

type Container interface {
	Content(func(Matcher))
}

type MatchIndexer interface {
	Matcher
	Indexer
}

func MatchIndexers(v []Matcher) ([]MatchIndexer, bool) {
	for _, m := range v {
		if _, ok := m.(MatchIndexer); !ok {
			return nil, false
		}
	}
	r := make([]MatchIndexer, len(v))
	for i := range r {
		r[i] = v[i].(MatchIndexer)
	}
	return r, true
}

type MatchIndexSizer interface {
	Matcher
	Indexer
	Sizer
}

func MatchIndexSizers(v []Matcher) ([]MatchIndexSizer, bool) {
	for _, m := range v {
		if _, ok := m.(MatchIndexSizer); !ok {
			return nil, false
		}
	}
	r := make([]MatchIndexSizer, len(v))
	for i := range r {
		r[i] = v[i].(MatchIndexSizer)
	}
	return r, true
}

func Graphviz(pattern string, m Matcher) string {
	return fmt.Sprintf(`digraph G {graph[label="%s"];%s}`, pattern, graphviz(m, fmt.Sprintf("%x", rand.Int63())))
}

func BuildMatcher(matchers []Matcher) (m Matcher, err error) {
	if debugEnabled {
		debugEnterPrefix("compiling %s", matchers)
		defer func() {
			debugLogf("-> %s, %v", m, err)
			debugLeavePrefix()
		}()
	}
	if len(matchers) == 0 {
		return nil, fmt.Errorf("compile error: need at least one matcher")
	}
	if len(matchers) == 1 {
		return matchers[0], nil
	}
	if m := glueMatchers(matchers); m != nil {
		return m, nil
	}
	var (
		x        = -1
		max      = -2
		wantText bool
		indexer  MatchIndexer
	)
	for i, m := range matchers {
		mx, ok := m.(MatchIndexer)
		if !ok {
			continue
		}
		_, isText := m.(TextMatcher)
		if wantText && !isText {
			continue
		}
		n := m.Len()
		if (!wantText && isText) || n > max {
			max = n
			x = i
			indexer = mx
			wantText = isText
		}
	}
	if indexer == nil {
		return nil, fmt.Errorf("can not index on matchers")
	}
	left := matchers[:x]
	var right []Matcher
	if len(matchers) > x+1 {
		right = matchers[x+1:]
	}
	var (
		l Matcher = NothingMatcher{}
		r Matcher = NothingMatcher{}
	)
	if len(left) > 0 {
		l, err = BuildMatcher(left)
		if err != nil {
			return nil, err
		}
	}
	if len(right) > 0 {
		r, err = BuildMatcher(right)
		if err != nil {
			return nil, err
		}
	}
	return NewTree(indexer, l, r), nil
}

func Optimize(m Matcher) (opt Matcher) {
	if debugEnabled {
		defer func() {
			a := fmt.Sprintf("%s", m)
			b := fmt.Sprintf("%s", opt)
			if a != b {
				debugEnterPrefix("optimized %s: -> %s", a, b)
				debugLeavePrefix()
			}
		}()
	}
	switch v := m.(type) {
	case AnyMatcher:
		if len(v.sep) == 0 {
			return NewSuper()
		}
	case ListMatcher:
		if v.not == false && len(v.rs) == 1 {
			return NewText(string(v.rs))
		}
		return m
	case TreeMatcher:
		v.left = Optimize(v.left)
		v.right = Optimize(v.right)
		txt, ok := v.value.(TextMatcher)
		if !ok {
			return m
		}
		var (
			leftNil  = v.left == nil || v.left == NothingMatcher{}
			rightNil = v.right == nil || v.right == NothingMatcher{}
		)
		if leftNil && rightNil {
			return NewText(txt.s)
		}
		_, leftSuper := v.left.(SuperMatcher)
		lp, leftPrefix := v.left.(PrefixMatcher)
		la, leftAny := v.left.(AnyMatcher)
		_, rightSuper := v.right.(SuperMatcher)
		rs, rightSuffix := v.right.(SuffixMatcher)
		ra, rightAny := v.right.(AnyMatcher)
		switch {
		case leftSuper && rightSuper:
			return NewContains(txt.s)
		case leftSuper && rightNil:
			return NewSuffix(txt.s)
		case rightSuper && leftNil:
			return NewPrefix(txt.s)
		case leftNil && rightSuffix:
			return NewPrefixSuffix(txt.s, rs.s)
		case rightNil && leftPrefix:
			return NewPrefixSuffix(lp.s, txt.s)
		case rightNil && leftAny:
			return NewSuffixAny(txt.s, la.sep)
		case leftNil && rightAny:
			return NewPrefixAny(txt.s, ra.sep)
		}
	case Container:
		var (
			first Matcher
			n     int
		)
		v.Content(func(m Matcher) {
			first = m
			n++
		})
		if n == 1 {
			return first
		}
		return m
	}
	return m
}

func glueMatchers(ms []Matcher) Matcher {
	if m := glueMatchersAsEvery(ms); m != nil {
		return m
	}
	if m := glueMatchersAsRow(ms); m != nil {
		return m
	}
	return nil
}

func glueMatchersAsRow(ms []Matcher) Matcher {
	if len(ms) <= 1 {
		return nil
	}
	var s []MatchIndexSizer
	for _, m := range ms {
		rsz, ok := m.(MatchIndexSizer)
		if !ok {
			return nil
		}
		s = append(s, rsz)
	}
	return NewRow(s)
}

func glueMatchersAsEvery(ms []Matcher) Matcher {
	if len(ms) <= 1 {
		return nil
	}
	var (
		hasAny    bool
		hasSuper  bool
		hasSingle bool
		minimum   int
		separator []rune
	)
	for i, matcher := range ms {
		var sep []rune
		switch m := matcher.(type) {
		case SuperMatcher:
			sep = []rune{}
			hasSuper = true
		case AnyMatcher:
			sep = m.sep
			hasAny = true
		case SingleMatcher:
			sep = m.sep
			hasSingle = true
			minimum++
		case ListMatcher:
			if !m.not {
				return nil
			}
			sep = m.rs
			hasSingle = true
			minimum++
		default:
			return nil
		}
		// initialize
		if i == 0 {
			separator = sep
		}
		if runesEqual(sep, separator) {
			continue
		}
		return nil
	}
	if hasSuper && !hasAny && !hasSingle {
		return NewSuper()
	}
	if hasAny && !hasSuper && !hasSingle {
		return NewAny(separator)
	}
	if (hasAny || hasSuper) && minimum > 0 && len(separator) == 0 {
		return NewMin(minimum)
	}
	var every []Matcher
	if minimum > 0 {
		every = append(every, NewMin(minimum))
		if !hasAny && !hasSuper {
			every = append(every, NewMax(minimum))
		}
	}
	if len(separator) > 0 {
		every = append(every, NewAny(separator))
	}
	return NewEveryOf(every)
}

type result struct {
	ms        []Matcher
	matchers  int
	maxMinLen int
	sumMinLen int
	nesting   int
}

func compareResult(a, b result) int {
	if x := b.sumMinLen - a.sumMinLen; x != 0 {
		return x
	}
	if x := len(a.ms) - len(b.ms); x != 0 {
		return x
	}
	if x := a.nesting - b.nesting; x != 0 {
		return x
	}
	if x := a.matchers - b.matchers; x != 0 {
		return x
	}
	if x := b.maxMinLen - a.maxMinLen; x != 0 {
		return x
	}
	return 0
}

func collapse(ms []Matcher, x Matcher, i, j int) (cp []Matcher) {
	cp = make([]Matcher, len(ms)-(j-i)+1)
	copy(cp[0:i], ms[0:i])
	copy(cp[i+1:], ms[j:])
	cp[i] = x
	return cp
}

func matchersCount(ms []Matcher) (n int) {
	n = len(ms)
	for _, m := range ms {
		n += countNestedMatchers(m)
	}
	return n
}

func countNestedMatchers(m Matcher) (n int) {
	if c, _ := m.(Container); c != nil {
		c.Content(func(m Matcher) {
			n += 1 + countNestedMatchers(m)
		})
	}
	return n
}

func nestingDepth(m Matcher) (depth int) {
	c, ok := m.(Container)
	if !ok {
		return 0
	}
	var max int
	c.Content(func(m Matcher) {
		if d := nestingDepth(m); d > max {
			max = d
		}
	})
	return max + 1
}

func maxMinLen(ms []Matcher) (max int) {
	for _, m := range ms {
		if n := m.Len(); n > max {
			max = n
		}
	}
	return max
}

func sumMinLen(ms []Matcher) (sum int) {
	for _, m := range ms {
		sum += m.Len()
	}
	return sum
}

func maxNestingDepth(ms []Matcher) (max int) {
	for _, m := range ms {
		if n := nestingDepth(m); n > max {
			max = n
		}
	}
	return
}

func minimizeMatcher(ms []Matcher, i, j int, best *result) *result {
	if j > len(ms) {
		j = 0
		i++
	}
	if i > len(ms)-2 {
		return best
	}
	if j == 0 {
		j = i + 2
	}
	if g := glueMatchers(ms[i:j]); g != nil {
		cp := collapse(ms, g, i, j)
		r := result{
			ms:        cp,
			matchers:  matchersCount(cp),
			sumMinLen: sumMinLen(cp),
			maxMinLen: maxMinLen(cp),
			nesting:   maxNestingDepth(cp),
		}
		if debugEnabled {
			debugEnterPrefix(
				"intermediate: %s (matchers:%d, summinlen:%d, maxminlen:%d, nesting:%d)",
				cp, r.matchers, r.sumMinLen, r.maxMinLen, r.nesting,
			)
		}
		if best == nil {
			best = new(result)
		}
		if best.ms == nil || compareResult(r, *best) < 0 {
			*best = r
			if debugEnabled {
				debugLogf("new best result")
			}
		}
		best = minimizeMatcher(cp, 0, 0, best)
		if debugEnabled {
			debugLeavePrefix()
		}
	}
	return minimizeMatcher(ms, i, j+1, best)
}

func MinimizeMatcher(ms []Matcher) (m []Matcher) {
	if debugEnabled {
		debugEnterPrefix("minimizing %s", ms)
		defer func() {
			debugLogf("-> %s", m)
			debugLeavePrefix()
		}()
	}
	best := minimizeMatcher(ms, 0, 0, nil)
	if best == nil {
		return ms
	}
	return best.ms
}

func graphviz(m Matcher, id string) string {
	buf := &bytes.Buffer{}
	switch v := m.(type) {
	case TreeMatcher:
		fmt.Fprintf(buf, `"%s"[label="%s"];`, id, v.value)
		for _, m := range []Matcher{v.left, v.right} {
			switch n := m.(type) {
			case nil:
				rnd := rand.Int63()
				fmt.Fprintf(buf, `"%x"[label="<nil>"];`, rnd)
				fmt.Fprintf(buf, `"%s"->"%x";`, id, rnd)
			default:
				sub := fmt.Sprintf("%x", rand.Int63())
				fmt.Fprintf(buf, `"%s"->"%s";`, id, sub)
				fmt.Fprint(buf, graphviz(n, sub))
			}
		}
	case Container:
		fmt.Fprintf(buf, `"%s"[label="Container(%T)"];`, id, m)
		v.Content(func(m Matcher) {
			rnd := rand.Int63()
			fmt.Fprint(buf, graphviz(m, fmt.Sprintf("%x", rnd)))
			fmt.Fprintf(buf, `"%s"->"%x";`, id, rnd)
		})
	default:
		fmt.Fprintf(buf, `"%s"[label="%s"];`, id, m)
	}
	return buf.String()
}

// appendMerge merges and sorts given already SORTED and UNIQUE segments.
func appendMerge(target, sub []int) []int {
	nt, ns := len(target), len(sub)
	v := make([]int, 0, nt+ns)
	for i, j := 0, 0; i < nt || j < ns; {
		if i >= nt {
			v = append(v, sub[j:]...)
			break
		}
		if j >= ns {
			v = append(v, target[i:]...)
			break
		}
		t, s := target[i], sub[j]
		switch {
		case t == s:
			v = append(v, t)
			i++
			j++
		case t < s:
			v = append(v, t)
			i++
		case s < t:
			v = append(v, s)
			j++
		}
	}
	return append(target[:0], v...)
}
