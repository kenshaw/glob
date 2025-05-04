package syntax

import (
	"fmt"

	"github.com/kenshaw/glob/debug"
	"github.com/kenshaw/glob/match"
)

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
