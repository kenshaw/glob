package syntax

import (
	"bytes"
	"fmt"

	"github.com/kenshaw/glob/debug"
	"github.com/kenshaw/glob/match"
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

// Match builds the matcher for the node.
func (node *Node) Match(sep []rune) (match.Matcher, error) {
	return buildMatch(node, sep)
}

func (node *Node) Equal(n *Node) bool {
	switch {
	case node.Type != n.Type,
		node.Value != n.Value,
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

func buildMatch(node *Node, sep []rune) (m match.Matcher, err error) {
	if debug.Enabled {
		debug.EnterPrefix("compiler: compiling %s", node)
		defer func() {
			if err != nil {
				debug.Logf("->! %v", err)
			} else {
				debug.Logf("-> %s", m)
			}
			debug.LeavePrefix()
		}()
	}
	// todo this could be faster on pattern_alternatives_combine_lite (see glob_test.go)
	if n := Minimize(node); n != nil {
		debug.Logf("minimized tree -> %s", node, n)
		r, err := buildMatch(n, sep)
		if debug.Enabled {
			if err != nil {
				debug.Logf("compiler: compile minimized tree failed: %v", err)
			} else {
				debug.Logf("compiler: minimized tree")
				debug.Logf("compiler: \t%s", node)
				debug.Logf("compiler: \t%s", n)
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
		return match.NewAnyOf(matchers...), nil
	case Pattern:
		if len(node.Children) == 0 {
			return match.NewNothing(), nil
		}
		matchers, err := buildNodeMatch(node.Children, sep)
		if err != nil {
			return nil, err
		}
		m, err = match.Compile(match.Minimize(matchers))
		if err != nil {
			return nil, err
		}
	case Any:
		m = match.NewAny(sep)
	case Super:
		m = match.NewSuper()
	case Single:
		m = match.NewSingle(sep)
	case Nothing:
		m = match.NewNothing()
	case List:
		l := node.Value.(ListData)
		m = match.NewList([]rune(l.Chars), l.Not)
	case Range:
		r := node.Value.(RangeData)
		m = match.NewRange(r.Lo, r.Hi, r.Not)
	case Text:
		t := node.Value.(TextData)
		m = match.NewText(t.Text)
	default:
		return nil, fmt.Errorf("could not compile tree: unknown node type %s (%d)", node.Type, int(node.Type))
	}
	return match.Optimize(m), nil
}

func buildNodeMatch(ns []*Node, sep []rune) ([]match.Matcher, error) {
	var matchers []match.Matcher
	for _, n := range ns {
		m, err := buildMatch(n, sep)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, m)
	}
	return matchers, nil
}
