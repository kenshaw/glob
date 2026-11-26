package syntax

import (
	"fmt"
	"math/bits"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"
)

// NFAMatcher matches a tree directly. It reads the input once for each node,
// and keeps the set of positions where the nodes so far can end, as the
// simulation of a nondeterministic automaton does. It matches every node
// type, including the extended groups that the optimized matchers cannot
// hold. It is also the reference that the tests compare the optimized
// matchers with (D12, D20).
//
// The time of a call is linear in the length of the input. A group that
// repeats, or a group !(...), whose patterns have no most width, such as
// +(*a), can take time that is quadratic in the length of the input. The
// depth of the recursion depends on the pattern alone, never on the input.
type NFAMatcher struct {
	root *nseq
	sep  []rune
	node *Node
}

// nseq is a sequence of nodes that must match one after the other.
type nseq struct {
	nodes []*nnode
}

// nnode is one node of a sequence.
type nnode struct {
	typ   Type
	text  string
	class func(rune) bool
	data  ClassData
	op    rune
	alts  []*nseq
	// ascii holds the answer of class for each ASCII rune.
	ascii [2]uint64
	// width is the most bytes that an alternative can match, or -1 when
	// there is no most.
	width int
}

// NewNFA creates an NFA matcher for the tree.
func NewNFA(node *Node, sep []rune) *NFAMatcher {
	return &NFAMatcher{
		root: compileSeq(node),
		sep:  slices.Clone(sep),
		node: node,
	}
}

// compileSeq turns a node of type Pattern, or any other node, into a
// sequence.
func compileSeq(node *Node) *nseq {
	s := new(nseq)
	if node.Type == Pattern {
		for _, c := range node.Children {
			s.nodes = append(s.nodes, compileNode(c))
		}
	} else {
		s.nodes = append(s.nodes, compileNode(node))
	}
	return s
}

// compileNode turns one node into an nnode.
func compileNode(node *Node) *nnode {
	n := &nnode{typ: node.Type}
	switch node.Type {
	case Text:
		n.text = node.Value.(TextData).Text
	case List:
		l := node.Value.(ListData)
		n.data = ClassData{Not: l.Not, Chars: l.Chars}
		n.typ = Class
	case Range:
		r := node.Value.(RangeData)
		n.data = ClassData{Not: r.Not, Ranges: []RangeData{r}}
		n.typ = Class
	case Class:
		n.data = node.Value.(ClassData)
	case Pattern:
		n.typ = AnyOf
		n.alts = []*nseq{compileSeq(node)}
	case Ext:
		n.op = node.Value.(ExtData).Op
		fallthrough
	case AnyOf:
		for _, c := range node.Children {
			n.alts = append(n.alts, compileSeq(c))
		}
	}
	if n.typ == Class {
		n.class, n.ascii = n.data.Contains, n.data.ASCII()
	}
	n.width = -1
	if n.alts != nil {
		n.width = 0
		for _, alt := range n.alts {
			w := alt.width()
			if w == -1 {
				n.width = -1
				break
			}
			n.width = max(n.width, w)
		}
	}
	return n
}

// width returns the most bytes that the sequence can match, or -1 when there
// is no most.
func (q *nseq) width() int {
	var sum int
	for _, n := range q.nodes {
		switch {
		case n.typ == Nothing:
		case n.typ == Text:
			sum += len(n.text)
		case n.typ == Single, n.typ == Class:
			sum += utf8.UTFMax
		case n.typ == AnyOf, n.typ == Ext && (n.op == charAt || n.op == charSingle):
			if n.width == -1 {
				return -1
			}
			sum += n.width
		default:
			return -1
		}
	}
	return sum
}

// Match satisfies the [Matcher] interface.
func (m *NFAMatcher) Match(s string) bool {
	st := getState(m, s)
	defer putState(st)
	in := st.alloc()
	in.set(0)
	out := st.seq(m.root, in)
	ok := out.has(len(s))
	st.release(out)
	return ok
}

// Len satisfies the [Matcher] interface. It returns 0, because the matcher
// does not compute a minimum length.
func (m *NFAMatcher) Len() int {
	return 0
}

// String satisfies the [fmt.Stringer] interface.
func (m *NFAMatcher) String() string {
	return fmt.Sprintf("<nfa:%s>", m.node)
}

// set is a set of positions in the input, one bit for each byte offset from
// 0 to the length of the input.
type set []uint64

// set adds the position i.
func (b set) set(i int) {
	b[i>>6] |= 1 << (uint(i) & 63)
}

// has reports whether the set holds the position i.
func (b set) has(i int) bool {
	return b[i>>6]&(1<<(uint(i)&63)) != 0
}

// empty reports whether the set holds no position.
func (b set) empty() bool {
	for _, w := range b {
		if w != 0 {
			return false
		}
	}
	return true
}

// count returns the number of positions in the set.
func (b set) count() int {
	var n int
	for _, w := range b {
		n += bits.OnesCount64(w)
	}
	return n
}

// first returns the smallest position of the set, or -1.
func (b set) first() int {
	for i, w := range b {
		if w != 0 {
			return i<<6 + bits.TrailingZeros64(w)
		}
	}
	return -1
}

// each calls f with each position of the set, from the smallest.
func (b set) each(f func(int)) {
	for i, w := range b {
		for w != 0 {
			f(i<<6 + bits.TrailingZeros64(w))
			w &= w - 1
		}
	}
}

// nstate holds the input and the scratch sets of one call of Match.
type nstate struct {
	m     *NFAMatcher
	s     string
	words int
	free  []set
}

// states holds the states that calls of Match reuse.
var states = sync.Pool{New: func() any { return new(nstate) }}

// getState returns a nstate for one call of Match.
func getState(m *NFAMatcher, s string) *nstate {
	st := states.Get().(*nstate)
	st.m = m
	st.reset(s)
	return st
}

// reset makes the state answer for the input s.
func (st *nstate) reset(s string) {
	st.s, st.words = s, len(s)>>6+1
}

// putState returns the nstate to the pool. The nstate keeps no input.
func putState(st *nstate) {
	st.m, st.s = nil, ""
	// A nstate that grew for a long input gives its memory back.
	if st.words > 1<<10 {
		st.free = nil
	}
	states.Put(st)
}

// alloc returns an empty set.
func (st *nstate) alloc() set {
	if n := len(st.free); n > 0 {
		b := st.free[n-1]
		st.free = st.free[:n-1]
		if cap(b) >= st.words {
			b = b[:st.words]
			clear(b)
			return b
		}
	}
	return make(set, st.words)
}

// release gives the set back for a later alloc.
func (st *nstate) release(b set) {
	st.free = append(st.free, b)
}

// isSep reports whether r is a separator.
func (st *nstate) isSep(r rune) bool {
	return slices.Contains(st.m.sep, r)
}

// seq returns the set of the positions where q ends, when q starts at each
// position of in. It releases in.
func (st *nstate) seq(q *nseq, in set) set {
	for _, n := range q.nodes {
		if in.empty() {
			return in
		}
		out := st.node(n, in)
		st.release(in)
		in = out
	}
	return in
}

// node returns the set of the positions where n ends, when n starts at each
// position of in. It does not change in.
func (st *nstate) node(n *nnode, in set) set {
	out, s := st.alloc(), st.s
	switch n.typ {
	case Nothing:
		copy(out, in)
	case Text:
		st.text(n.text, in, out)
	case Single:
		in.each(func(p int) {
			if p == len(s) {
				return
			}
			if r, w := utf8.DecodeRuneInString(s[p:]); !st.isSep(r) {
				out.set(p + w)
			}
		})
	case Class:
		in.each(func(p int) {
			if p == len(s) {
				return
			}
			if c := s[p]; c < utf8.RuneSelf {
				if n.ascii[c>>6]&(1<<(c&63)) != 0 {
					out.set(p + 1)
				}
				return
			}
			if r, w := utf8.DecodeRuneInString(s[p:]); n.class(r) {
				out.set(p + w)
			}
		})
	case Super:
		for p := in.first(); ; {
			out.set(p)
			if p == len(s) {
				break
			}
			_, w := utf8.DecodeRuneInString(s[p:])
			p += w
		}
	case Any:
		// One pass from the first position. A star that starts at a position
		// of in runs until the next separator.
		for p, on := in.first(), false; ; {
			on = on || in.has(p)
			if on {
				out.set(p)
			}
			if p == len(s) {
				break
			}
			r, w := utf8.DecodeRuneInString(s[p:])
			if st.isSep(r) {
				on = false
			}
			p += w
		}
	case AnyOf:
		st.alts(n, in, out)
	case Ext:
		switch n.op {
		case charAt:
			st.alts(n, in, out)
		case charSingle:
			copy(out, in)
			st.alts(n, in, out)
		case charPlus:
			st.repeat(n, in, out)
		case charAny:
			copy(out, in)
			st.repeat(n, in, out)
		case charRangeNot:
			st.none(n, in, out)
		}
	}
	return out
}

// text adds to out the end of text, for each position of in where text
// starts. When in holds many positions, as after a star, it finds each
// occurrence of text with [strings.Index], and does not test each position.
func (st *nstate) text(text string, in, out set) {
	s := st.s
	if in.count() <= 4 {
		in.each(func(p int) {
			if strings.HasPrefix(s[p:], text) {
				out.set(p + len(text))
			}
		})
		return
	}
	for p := in.first(); p <= len(s)-len(text); p++ {
		i := strings.Index(s[p:], text)
		if i == -1 {
			return
		}
		p += i
		if in.has(p) {
			out.set(p + len(text))
		}
	}
}

// alts adds to out the positions where an alternative of n ends, when it
// starts at a position of in.
func (st *nstate) alts(n *nnode, in, out set) {
	for _, alt := range n.alts {
		c := st.alloc()
		copy(c, in)
		c = st.seq(alt, c)
		for i, w := range c {
			out[i] |= w
		}
		st.release(c)
	}
}

// repeat adds to out the positions where one or more of the alternatives of
// n, one after the other, end, when they start at a position of in.
func (st *nstate) repeat(n *nnode, in, out set) {
	if n.width != -1 {
		st.repeatSweep(n, in, out)
		return
	}
	// The alternatives have no most width. Add the ends of the newest
	// positions until no position is new. The time is quadratic at worst.
	seen, next := st.alloc(), st.alloc()
	st.alts(n, in, seen)
	frontier := st.alloc()
	copy(frontier, seen)
	for !frontier.empty() {
		clear(next)
		st.alts(n, frontier, next)
		for i, w := range next {
			frontier[i] = w &^ seen[i]
			seen[i] |= w
		}
	}
	for i, w := range seen {
		out[i] |= w
	}
	st.release(seen)
	st.release(next)
	st.release(frontier)
}

// repeatSweep answers repeat when the alternatives match at most n.width
// bytes. A repetition ends at or after the position where it starts, so one
// pass from the first position finds each end. Each position runs the
// alternatives on a window of n.width bytes, so the time is linear in the
// length of the input.
func (st *nstate) repeatSweep(n *nnode, in, out set) {
	s, ends := st.s, st.alloc()
	sub := getState(st.m, "")
	for p := in.first(); p <= len(s); p++ {
		if !in.has(p) && !ends.has(p) {
			continue
		}
		sub.reset(s[p:min(len(s), p+n.width)])
		one, e := sub.alloc(), sub.alloc()
		one.set(0)
		sub.alts(n, one, e)
		e.each(func(i int) { ends.set(p + i) })
		sub.release(one)
		sub.release(e)
	}
	putState(sub)
	for i, w := range ends {
		out[i] |= w
	}
	st.release(ends)
}

// none adds to out the positions e where no alternative of n matches s[p:e]
// for a position p of in, and s[p:e] holds no separator. A group !(...)
// matches inside one segment, as * does.
//
// When the alternatives match at most n.width bytes, each e after p+width
// is in out, up to the end of the segment. A pass as for a star adds those
// positions, so the time is linear in the length of the input. Otherwise the
// time is quadratic at worst.
func (st *nstate) none(n *nnode, in, out set) {
	s, tails := st.s, st.alloc()
	sub := getState(st.m, "")
	in.each(func(p int) {
		// The alternatives run on a window of the input that starts at p.
		// They cannot match more than n.width bytes, so a window of that
		// width keeps the time of each position bounded.
		hi := len(s)
		if n.width != -1 {
			hi = min(hi, p+n.width)
		}
		sub.reset(s[p:hi])
		one, ends := sub.alloc(), sub.alloc()
		one.set(0)
		sub.alts(n, one, ends)
		for e := p; ; {
			if e > hi {
				// No alternative reaches e, nor any later position.
				tails.set(e)
				break
			}
			if !ends.has(e - p) {
				out.set(e)
			}
			if e == len(s) {
				break
			}
			r, w := utf8.DecodeRuneInString(s[e:])
			if st.isSep(r) {
				break
			}
			e += w
		}
		sub.release(one)
		sub.release(ends)
	})
	putState(sub)
	if p := tails.first(); p != -1 {
		for on := false; ; {
			on = on || tails.has(p)
			if on {
				out.set(p)
			}
			if p == len(s) {
				break
			}
			r, w := utf8.DecodeRuneInString(s[p:])
			if st.isSep(r) {
				on = false
			}
			p += w
		}
	}
	st.release(tails)
}
