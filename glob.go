// Package glob provides glob matching.
package glob

import (
	"fmt"

	"github.com/kenshaw/glob/debug"
	"github.com/kenshaw/glob/match"
	"github.com/kenshaw/glob/syntax"
	"github.com/kenshaw/glob/syntax/ast"
)

// Glob represents compiled glob pattern.
type Glob interface {
	Match(string) bool
}

// Compile creates Glob for given pattern and strings (if any present after
// pattern) as separators. The pattern syntax is:
//
//	pattern:
//	    { term }
//
//	term:
//	    `*`         matches any sequence of non-separator characters
//	    `**`        matches any sequence of characters
//	    `?`         matches any single non-separator character
//	    `[` [ `!` ] { character-range } `]`
//	                character class (must be non-empty)
//	    `{` pattern-list `}`
//	                pattern alternatives
//	    c           matches character c (c != `*`, `**`, `?`, `\`, `[`, `{`, `}`)
//	    `\` c       matches character c
//
//	character-range:
//	    c           matches character c (c != `\\`, `-`, `]`)
//	    `\` c       matches character c
//	    lo `-` hi   matches character c for lo <= c <= hi
//
//	pattern-list:
//	    pattern { `,` pattern }
//	                comma-separated (without spaces) patterns
func Compile(pattern string, separators ...rune) (Glob, error) {
	ast, err := syntax.Parse(pattern)
	if err != nil {
		return nil, err
	}
	matcher, err := CompileTree(ast, separators)
	if err != nil {
		return nil, err
	}
	return matcher, nil
}

// MustCompile is the same as Compile, except that if Compile returns error,
// this will panic
func MustCompile(pattern string, separators ...rune) Glob {
	g, err := Compile(pattern, separators...)
	if err != nil {
		panic(err)
	}
	return g
}

// QuoteMeta returns a string that quotes all glob pattern meta characters
// inside the argument text; For example, QuoteMeta(`{foo*}`) returns `\[foo\*\]`.
func QuoteMeta(s string) string {
	b := make([]byte, 2*len(s))
	// a byte loop is correct because all meta characters are ASCII
	j := 0
	for i := range len(s) {
		if syntax.IsSpecial(s[i]) {
			b[j] = '\\'
			j++
		}
		b[j] = s[i]
		j++
	}
	return string(b[0:j])
}

// TODO use constructor with all matchers, and to their structs private
// TODO glue multiple Text nodes (like after QuoteMeta)

func CompileTree(tree *ast.Node, sep []rune) (match.Matcher, error) {
	m, err := compileNode(tree, sep)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func compileNode(node *ast.Node, sep []rune) (m match.Matcher, err error) {
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
	if n := ast.Minimize(node); n != nil {
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
	case ast.KindAnyOf:
		matchers, err := compileNodes(node.Children, sep)
		if err != nil {
			return nil, err
		}
		return match.NewAnyOf(matchers...), nil
	case ast.KindPattern:
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
	case ast.KindAny:
		m = match.NewAny(sep)
	case ast.KindSuper:
		m = match.NewSuper()
	case ast.KindSingle:
		m = match.NewSingle(sep)
	case ast.KindNothing:
		m = match.NewNothing()
	case ast.KindList:
		l := node.Value.(ast.List)
		m = match.NewList([]rune(l.Chars), l.Not)
	case ast.KindRange:
		r := node.Value.(ast.Range)
		m = match.NewRange(r.Lo, r.Hi, r.Not)
	case ast.KindText:
		t := node.Value.(ast.Text)
		m = match.NewText(t.Text)
	default:
		return nil, fmt.Errorf("could not compile tree: unknown node type %s (%d)", node.Kind, int(node.Kind))
	}
	return match.Optimize(m), nil
}

func compileNodes(ns []*ast.Node, sep []rune) ([]match.Matcher, error) {
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
