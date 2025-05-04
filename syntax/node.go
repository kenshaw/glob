package syntax

import (
	"bytes"
	"fmt"

	"github.com/kenshaw/glob/debug"
	"github.com/kenshaw/glob/match"
)

type Kind int

const (
	KindNothing Kind = iota
	KindPattern
	KindList
	KindRange
	KindText
	KindAny
	KindSuper
	KindSingle
	KindAnyOf
)

func (kind Kind) String() string {
	switch kind {
	case KindNothing:
		return "Nothing"
	case KindPattern:
		return "Pattern"
	case KindList:
		return "List"
	case KindRange:
		return "Range"
	case KindText:
		return "Text"
	case KindAny:
		return "Any"
	case KindSuper:
		return "Super"
	case KindSingle:
		return "Single"
	case KindAnyOf:
		return "AnyOf"
	default:
		return ""
	}
}

type Node struct {
	Parent   *Node
	Kind     Kind
	Value    any
	Children []*Node
}

func New(k Kind, v any, ch ...*Node) *Node {
	node := &Node{
		Kind:  k,
		Value: v,
	}
	for _, c := range ch {
		node.Insert(c)
	}
	return node
}

func (node *Node) Insert(children ...*Node) {
	node.Children = append(node.Children, children...)
	for _, ch := range children {
		ch.Parent = node
	}
}

func (node *Node) Compile(sep []rune) (match.Matcher, error) {
	return compileNode(node, sep)
}

func (node *Node) Equal(n *Node) bool {
	switch {
	case node.Kind != n.Kind,
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
	buf.WriteString(node.Kind.String())
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

type List struct {
	Not   bool
	Chars string
}

type Range struct {
	Not    bool
	Lo, Hi rune
}

type Text struct {
	Text string
}

// TODO use constructor with all matchers, and to their structs private
// TODO glue multiple Text nodes (like after QuoteMeta)

func compileNode(node *Node, sep []rune) (m match.Matcher, err error) {
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
		r, err := compileNode(n, sep)
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
	switch node.Kind {
	case KindAnyOf:
		matchers, err := compileNodes(node.Children, sep)
		if err != nil {
			return nil, err
		}
		return match.NewAnyOf(matchers...), nil
	case KindPattern:
		if len(node.Children) == 0 {
			return match.NewNothing(), nil
		}
		matchers, err := compileNodes(node.Children, sep)
		if err != nil {
			return nil, err
		}
		m, err = match.Compile(match.Minimize(matchers))
		if err != nil {
			return nil, err
		}
	case KindAny:
		m = match.NewAny(sep)
	case KindSuper:
		m = match.NewSuper()
	case KindSingle:
		m = match.NewSingle(sep)
	case KindNothing:
		m = match.NewNothing()
	case KindList:
		l := node.Value.(List)
		m = match.NewList([]rune(l.Chars), l.Not)
	case KindRange:
		r := node.Value.(Range)
		m = match.NewRange(r.Lo, r.Hi, r.Not)
	case KindText:
		t := node.Value.(Text)
		m = match.NewText(t.Text)
	default:
		return nil, fmt.Errorf("could not compile tree: unknown node type %s (%d)", node.Kind, int(node.Kind))
	}
	return match.Optimize(m), nil
}

func compileNodes(ns []*Node, sep []rune) ([]match.Matcher, error) {
	var matchers []match.Matcher
	for _, n := range ns {
		m, err := compileNode(n, sep)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, m)
	}
	return matchers, nil
}
