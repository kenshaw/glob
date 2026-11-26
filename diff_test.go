package glob

import (
	"testing"
	"unicode/utf8"

	"github.com/kenshaw/glob/syntax"
)

// backtrack compiles the pattern with the backtracking matcher, which
// matches the tree without the optimizer. It is the reference that the
// optimized matchers must agree with.
func backtrack(tb testing.TB, pattern string, sep []rune) (syntax.Matcher, bool) {
	tb.Helper()
	tree, err := syntax.Parse(syntax.NewLexer(pattern))
	if err != nil {
		return nil, false
	}
	return syntax.NewBacktrack(tree, sep), true
}

// FuzzMatchersAgree compares the optimized matchers with the backtracking
// matcher on random patterns and inputs.
func FuzzMatchersAgree(f *testing.F) {
	for _, seed := range []struct {
		pattern, s string
		sep        bool
	}{
		{"a*ant", "ant", false},
		{"{*.,}google*", "a.google.com", true},
		{"{a,}{a,}a", "a", false},
		{"[a-z][!a-x]*cat*[h][!b]*eyes*", "my cat has very bright eyes", false},
		{"{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}", "http://yahoo.com", true},
		{"a.**.b", "a.b", true},
		{"*?*", "", false},
		{"[[:alpha:]]?[^0-9]", "ab_", false},
		{"x{1..12}y", "x10y", false},
		{"**", "a.b", true},
		{"?", "", false},
	} {
		f.Add(seed.pattern, seed.s, seed.sep)
	}
	f.Fuzz(func(t *testing.T, pattern, s string, withSep bool) {
		if !utf8.ValidString(pattern) || !utf8.ValidString(s) || len(pattern) > 64 || len(s) > 64 {
			t.Skip()
		}
		var sep []rune
		if withSep {
			sep = []rune{'.'}
		}
		g, err := Compile(pattern, sep...)
		ref, ok := backtrack(t, pattern, sep)
		switch {
		case err != nil && ok:
			t.Fatalf("Compile(%q) failed and Parse did not: %v", pattern, err)
		case err != nil:
			return
		}
		if got, want := g.Match(s), ref.Match(s); got != want {
			t.Fatalf("pattern %q, input %q, separators %q: expected %t, got: %t (matcher %s)", pattern, s, string(sep), want, got, g.Matcher)
		}
	})
}

// FuzzQuote holds D15. A glob compiled from Quote(s) matches s, alone and
// inside a group, and does not match a string that differs from s.
func FuzzQuote(f *testing.F) {
	for _, seed := range []string{"", "a", "*?[]{}", "a,b", "@(a|b)", `\`, "!(x)", "{1..3}", "[[:alpha:]]"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			t.Skip()
		}
		q := Quote(s)
		for _, pattern := range []string{q, "{" + q + ",\x00}", "@(" + q + "|\x00)"} {
			g, err := Compile(pattern)
			if err != nil {
				t.Fatalf("Compile(%q): expected no error, got: %v", pattern, err)
			}
			if !g.Match(s) {
				t.Errorf("Compile(%q) does not match %q", pattern, s)
			}
			if s != "x" && g.Match("x") {
				t.Errorf("Compile(%q) matches %q", pattern, "x")
			}
			if g.Match(s + "x") {
				t.Errorf("Compile(%q) matches %q", pattern, s+"x")
			}
		}
	})
}
