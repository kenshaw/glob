package syntax

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

type Type int

const (
	Nothing Type = iota
	Pattern
	List
	Range
	Text
	Any
	Super
	Single
	AnyOf
	Class
	Ext
)

func (typ Type) String() string {
	switch typ {
	case Nothing:
		return "Nothing"
	case Pattern:
		return "Pattern"
	case List:
		return "List"
	case Range:
		return "Range"
	case Text:
		return "Text"
	case Any:
		return "Any"
	case Super:
		return "Super"
	case Single:
		return "Single"
	case AnyOf:
		return "AnyOf"
	case Class:
		return "Class"
	case Ext:
		return "Ext"
	}
	return ""
}

type Node struct {
	Parent   *Node
	Type     Type
	Value    any
	Children []*Node
}

// New creates a new node of type.
func New(typ Type, v any, ch ...*Node) *Node {
	node := &Node{
		Type:  typ,
		Value: v,
	}
	for _, c := range ch {
		node.Insert(c)
	}
	return node
}

// Insert inserts a child node.
func (node *Node) Insert(children ...*Node) {
	node.Children = append(node.Children, children...)
	for _, ch := range children {
		ch.Parent = node
	}
}

// Match builds the matcher for the node. A tree gets the matchers that the
// compiler optimizes when they take time linear in the length of the input
// (see [Linear]). Every other tree gets an [NFAMatcher]: a tree that holds an
// extended group, a tree that the compiler cannot build, and a tree whose
// optimized matchers can take more than linear time (D20).
func (node *Node) Match(sep []rune) (Matcher, error) {
	if node.Contains(Ext) {
		return NewNFA(node, sep), nil
	}
	m, err := buildMatch(node, sep)
	if err != nil || !Linear(m) {
		// The optimizer cannot build every tree, such as a group that holds
		// only empty groups and stars. The NFA matcher matches any tree that
		// the parser builds.
		return NewNFA(node, sep), nil //nolint:nilerr // the fallback answers the error
	}
	return m, nil
}

// Contains reports whether the node or one of its descendants is of the type.
func (node *Node) Contains(typ Type) bool {
	if node.Type == typ {
		return true
	}
	for _, c := range node.Children {
		if c.Contains(typ) {
			return true
		}
	}
	return false
}

func (node *Node) Equal(n *Node) bool {
	switch {
	case node.Type != n.Type,
		!reflect.DeepEqual(node.Value, n.Value),
		len(node.Children) != len(n.Children):
		return false
	}
	for i, c := range node.Children {
		if !c.Equal(n.Children[i]) {
			return false
		}
	}
	return true
}

func (node *Node) String() string {
	var buf bytes.Buffer
	buf.WriteString(node.Type.String())
	if node.Value != nil {
		buf.WriteString(" =")
		buf.WriteString(fmt.Sprintf("%v", node.Value))
	}
	if len(node.Children) > 0 {
		buf.WriteString(" [")
		for i, c := range node.Children {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(c.String())
		}
		buf.WriteString("]")
	}
	return buf.String()
}

type ListData struct {
	Not   bool
	Chars string
}

type RangeData struct {
	Not    bool
	Lo, Hi rune
}

type TextData struct {
	Text string
}

// TODO use constructor with all matchers, and to their structs private
// TODO glue multiple Text nodes (like after QuoteMeta)

func buildMatch(node *Node, sep []rune) (m Matcher, err error) {
	if debugEnabled {
		debugEnterPrefix("compiler: compiling %s", node)
		defer func() {
			if err != nil {
				debugLogf("->! %v", err)
			} else {
				debugLogf("-> %s", m)
			}
			debugLeavePrefix()
		}()
	}
	// todo this could be faster on pattern_alternatives_combine_lite (see glob_test.go)
	if n := Minimize(node); n != nil {
		debugLogf("minimized tree -> %s", node, n)
		r, err := buildMatch(n, sep)
		if debugEnabled {
			if err != nil {
				debugLogf("compiler: compile minimized tree failed: %v", err)
			} else {
				debugLogf("compiler: minimized tree")
				debugLogf("compiler: \t%s", node)
				debugLogf("compiler: \t%s", n)
			}
		}
		if err == nil {
			return r, nil
		}
	}
	switch node.Type {
	case AnyOf:
		matchers, err := buildNodeMatch(node.Children, sep)
		if err != nil {
			return nil, err
		}
		return NewAnyOf(matchers...), nil
	case Pattern:
		if len(node.Children) == 0 {
			return NewNothing(), nil
		}
		matchers, err := buildNodeMatch(squashStars(node.Children), sep)
		if err != nil {
			return nil, err
		}
		m, err = BuildMatcher(MinimizeMatcher(matchers))
		if err != nil {
			return nil, err
		}
	case Any:
		m = NewAny(sep)
	case Super:
		m = NewSuper()
	case Single:
		m = NewSingle(sep)
	case Nothing:
		m = NewNothing()
	case List:
		l := node.Value.(ListData)
		m = NewList([]rune(l.Chars), l.Not)
	case Range:
		r := node.Value.(RangeData)
		m = NewRange(r.Lo, r.Hi, r.Not)
	case Text:
		t := node.Value.(TextData)
		m = NewText(t.Text)
	case Class:
		m = NewClass(node.Value.(ClassData))
	case Ext:
		return nil, errors.New("could not compile tree: an extended group needs the NFA matcher")
	default:
		return nil, fmt.Errorf("could not compile tree: unknown node type %s (%d)", node.Type, int(node.Type))
	}
	return Optimize(m), nil
}

func buildNodeMatch(ns []*Node, sep []rune) ([]Matcher, error) {
	var matchers []Matcher
	for _, n := range ns {
		m, err := buildMatch(n, sep)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, m)
	}
	return matchers, nil
}

// squashStars returns the nodes with each run of stars as one star. A run
// that holds a super star is a super star, because ** matches all that *
// matches.
func squashStars(nodes []*Node) []*Node {
	out := make([]*Node, 0, len(nodes))
	for _, n := range nodes {
		if len(out) != 0 && (n.Type == Any || n.Type == Super) {
			if last := out[len(out)-1]; last.Type == Any || last.Type == Super {
				if n.Type == Super {
					out[len(out)-1] = n
				}
				continue
			}
		}
		out = append(out, n)
	}
	return out
}
