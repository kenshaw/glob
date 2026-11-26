package syntax

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

// BacktrackMatcher matches a tree directly, by trying each way to split the
// input. It matches every node type, including the extended groups that the
// optimized matchers cannot hold. It remembers each answer for the length of
// one call, so a call takes at most cubic time in the length of the input.
type BacktrackMatcher struct {
	root *seq
	sep  []rune
	seqs int
	src  string
}

// seq is a sequence of nodes that must match one after the other.
type seq struct {
	id    int
	nodes []*bnode
}

// bnode is one node of a sequence.
type bnode struct {
	typ   Type
	text  string
	class func(rune) bool
	op    rune
	alts  []*seq
}

// NewBacktrack creates a backtracking matcher for the tree.
func NewBacktrack(node *Node, sep []rune) *BacktrackMatcher {
	m := &BacktrackMatcher{sep: sep, src: node.String()}
	m.root = m.compile(node)
	return m
}

// compile turns a node of type Pattern, or any other node, into a sequence.
func (m *BacktrackMatcher) compile(node *Node) *seq {
	s := &seq{id: m.seqs}
	m.seqs++
	if node.Type == Pattern {
		for _, c := range node.Children {
			s.nodes = append(s.nodes, m.compileNode(c))
		}
	} else {
		s.nodes = append(s.nodes, m.compileNode(node))
	}
	return s
}

// compileNode turns one node into a bnode.
func (m *BacktrackMatcher) compileNode(node *Node) *bnode {
	n := &bnode{typ: node.Type}
	switch node.Type {
	case Text:
		n.text = node.Value.(TextData).Text
	case List:
		l := node.Value.(ListData)
		n.class = ClassData{Not: l.Not, Chars: l.Chars}.Contains
		n.typ = Class
	case Range:
		r := node.Value.(RangeData)
		n.class = ClassData{Not: r.Not, Ranges: []RangeData{r}}.Contains
		n.typ = Class
	case Class:
		n.class = node.Value.(ClassData).Contains
	case Ext:
		n.op = node.Value.(ExtData).Op
		fallthrough
	case AnyOf, Pattern:
		if node.Type == Pattern {
			n.typ = AnyOf
			n.alts = []*seq{m.compile(node)}
			break
		}
		for _, c := range node.Children {
			n.alts = append(n.alts, m.compile(c))
		}
	}
	return n
}

// Match satisfies the [Matcher] interface.
func (m *BacktrackMatcher) Match(s string) bool {
	st := &state{m: m, s: s, memo: make(map[key]bool), reps: make(map[rkey]bool)}
	return st.match(m.root, 0, 0, len(s))
}

// Len satisfies the [Matcher] interface. It returns 0, because the matcher
// does not compute a minimum length.
func (m *BacktrackMatcher) Len() int {
	return 0
}

// String satisfies the [fmt.Stringer] interface.
func (m *BacktrackMatcher) String() string {
	return fmt.Sprintf("<backtrack:%s>", m.src)
}

// key identifies one question: does the sequence id, from its node k, match
// the input from i to j.
type key struct {
	id, k, i, j int
}

// rkey identifies one question of repeat: do the alternatives of n, one or
// more times, match the input from i to j.
type rkey struct {
	n    *bnode
	i, j int
}

// state holds the input and the answers of one call of Match.
type state struct {
	m    *BacktrackMatcher
	s    string
	memo map[key]bool
	reps map[rkey]bool
}

// isSep reports whether r is a separator.
func (st *state) isSep(r rune) bool {
	return slices.Contains(st.m.sep, r)
}

// match reports whether the nodes of q, from the node k, match s[i:j].
func (st *state) match(q *seq, k, i, j int) bool {
	if k == len(q.nodes) {
		return i == j
	}
	key := key{q.id, k, i, j}
	if v, ok := st.memo[key]; ok {
		return v
	}
	v := st.matchNode(q, k, i, j)
	st.memo[key] = v
	return v
}

// matchNode answers match for the node k of q.
func (st *state) matchNode(q *seq, k, i, j int) bool {
	n, s := q.nodes[k], st.s
	switch n.typ {
	case Nothing:
		return st.match(q, k+1, i, j)
	case Text:
		if !strings.HasPrefix(s[i:j], n.text) {
			return false
		}
		return st.match(q, k+1, i+len(n.text), j)
	case Single, Class:
		if i == j {
			return false
		}
		r, w := utf8.DecodeRuneInString(s[i:j])
		if n.typ == Single && st.isSep(r) || n.typ == Class && !n.class(r) {
			return false
		}
		return st.match(q, k+1, i+w, j)
	case Any, Super:
		for e := i; ; {
			if st.match(q, k+1, e, j) {
				return true
			}
			if e == j {
				return false
			}
			r, w := utf8.DecodeRuneInString(s[e:j])
			if n.typ == Any && st.isSep(r) {
				return false
			}
			e += w
		}
	case AnyOf:
		return st.split(q, k, i, j, func(e int) bool { return st.any(n, i, e) })
	case Ext:
		switch n.op {
		case charAt:
			return st.split(q, k, i, j, func(e int) bool { return st.any(n, i, e) })
		case charSingle:
			return st.match(q, k+1, i, j) ||
				st.split(q, k, i, j, func(e int) bool { return st.any(n, i, e) })
		case charAny:
			return st.star(q, k, i, j)
		case charPlus:
			return st.split(q, k, i, j, func(e int) bool { return st.repeat(n, i, e) })
		case charRangeNot:
			return st.split(q, k, i, j, func(e int) bool { return st.none(n, i, e) })
		}
	}
	return false
}

// split reports whether some end e, from i to j, gives part(e) and a match
// of the rest of q from e to j.
func (st *state) split(q *seq, k, i, j int, part func(int) bool) bool {
	for e := i; ; {
		if part(e) && st.match(q, k+1, e, j) {
			return true
		}
		if e == j {
			return false
		}
		_, w := utf8.DecodeRuneInString(st.s[e:j])
		e += w
	}
}

// any reports whether an alternative of n matches s[i:e].
func (st *state) any(n *bnode, i, e int) bool {
	for _, alt := range n.alts {
		if st.match(alt, 0, i, e) {
			return true
		}
	}
	return false
}

// repeat reports whether one or more alternatives of n, one after the other,
// match s[i:e]. Each repetition matches at least one rune.
func (st *state) repeat(n *bnode, i, e int) bool {
	key := rkey{n, i, e}
	if v, ok := st.reps[key]; ok {
		return v
	}
	v := st.any(n, i, e)
	for m := i; !v && m < e; {
		_, w := utf8.DecodeRuneInString(st.s[m:e])
		m += w
		v = m < e && st.any(n, i, m) && st.repeat(n, m, e)
	}
	st.reps[key] = v
	return v
}

// star answers match for a group *(...) at the node k of q.
func (st *state) star(q *seq, k, i, j int) bool {
	return st.match(q, k+1, i, j) ||
		st.split(q, k, i, j, func(e int) bool { return e > i && st.repeat(q.nodes[k], i, e) })
}

// none reports whether no alternative of n matches s[i:e], and s[i:e] holds
// no separator. A group !(...) matches inside one segment, as * does.
func (st *state) none(n *bnode, i, e int) bool {
	if strings.ContainsFunc(st.s[i:e], st.isSep) {
		return false
	}
	return !st.any(n, i, e)
}
