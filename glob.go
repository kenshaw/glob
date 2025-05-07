// Package glob provides glob matching.
package glob

import (
	"fmt"

	"github.com/kenshaw/glob/syntax"
)

// Glob matches glob patterns.
type Glob struct {
	syntax.Matcher
}

// String satisfies the [fmt.Stringer] interface.
func (g *Glob) String() string {
	return fmt.Sprintf("%v", g.Matcher)
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
func Compile(pattern string, separators ...rune) (*Glob, error) {
	tree, err := syntax.Parse(syntax.NewLexer(pattern))
	if err != nil {
		return nil, err
	}
	m, err := tree.Match(separators)
	if err != nil {
		return nil, err
	}
	return &Glob{m}, nil
}

// Must is the same as Compile, except that if Compile returns error, this will
// panic
func Must(pattern string, separators ...rune) *Glob {
	g, err := Compile(pattern, separators...)
	if err != nil {
		panic(err)
	}
	return g
}

// Quote returns a string that quotes all glob pattern meta characters
// inside the argument text; For example, Quote(`{foo*}`) returns
// `\[foo\*\]`.
func Quote(s string) string {
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
