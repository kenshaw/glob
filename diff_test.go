package glob

import (
	"testing"
	"unicode/utf8"

	"github.com/kenshaw/glob/syntax"
)

// reference compiles the pattern with the NFA matcher, which matches the
// tree without the optimizer. It is the reference that the optimized
// matchers must agree with. It changes the tree as the options say, as
// CompileWith does.
func reference(tb testing.TB, pattern string, opts ...Option) (*Glob, bool) {
	tb.Helper()
	g := New(opts...)
	tree, sep, err := g.opts.tree(pattern)
	if err != nil {
		return nil, false
	}
	g.Matcher, g.pattern = syntax.NewNFA(tree, sep), pattern
	return g, true
}

// fuzzSeparators are the sets of separators that the fuzz tests choose from.
var fuzzSeparators = [][]rune{nil, {'.'}, {'.', '/'}, {'é', '.'}}

// fuzzOptions returns the options that the fuzz tests choose with sep and
// flags.
func fuzzOptions(sep, flags uint8) []Option {
	opts := []Option{WithSeparators(fuzzSeparators[int(sep)%len(fuzzSeparators)]...)}
	if flags&1 != 0 {
		opts = append(opts, WithFold())
	}
	if flags&2 != 0 {
		opts = append(opts, WithGlobstar())
	}
	return opts
}

// FuzzMatchersAgree compares the optimized matchers with the NFA
// matcher on random patterns and inputs.
func FuzzMatchersAgree(f *testing.F) {
	for _, seed := range []struct {
		pattern, s string
		sep, flags uint8
	}{
		{"a*ant", "ant", 0, 0},
		{"{*.,}google*", "a.google.com", 1, 0},
		{"{a,}{a,}a", "a", 0, 0},
		{"[a-z][!a-x]*cat*[h][!b]*eyes*", "my cat has very bright eyes", 0, 0},
		{"{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}", "http://yahoo.com", 1, 0},
		{"a.**.b", "a.b", 1, 2},
		{"*?*", "", 0, 0},
		{"[[:alpha:]]?[^0-9]", "ab_", 0, 1},
		{"x{1..12}y", "x10y", 0, 0},
		{"**", "a.b", 1, 0},
		{"?", "", 0, 0},
		{"*a*b", "a/a.b", 2, 0},
		{"*É*", "xéy.é", 3, 1},
		{"a/**/b", "a/b", 2, 2},
	} {
		f.Add(seed.pattern, seed.s, seed.sep, seed.flags)
	}
	f.Fuzz(func(t *testing.T, pattern, s string, sep, flags uint8) {
		if !utf8.ValidString(pattern) || !utf8.ValidString(s) || len(pattern) > 64 || len(s) > 64 {
			t.Skip()
		}
		opts := fuzzOptions(sep, flags)
		g, err := CompileWith(pattern, opts...)
		ref, ok := reference(t, pattern, opts...)
		switch {
		case err != nil && ok:
			t.Fatalf("Compile(%q) failed and Parse did not: %v", pattern, err)
		case err != nil:
			return
		}
		if got, want := g.Match(s), ref.Match(s); got != want {
			t.Fatalf("pattern %q, input %q, separators %q, flags %d: expected %t, got: %t (matcher %s)", pattern, s, string(g.Separators()), flags, want, got, g.Matcher)
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

// FuzzCompile holds the safety of the API on any input: the bytes of a
// pattern that is not UTF-8, a long pattern, and each option. Compile
// returns an error or a glob, and never panics. A glob matches any input
// without a panic, and its String compiles to a glob that matches the same.
func FuzzCompile(f *testing.F) {
	for _, seed := range []string{"", "*", "\\", "[", "{", "@(", "[[:", "{1..3}", "\xff*", "a/**/b", "[\\]-a]", "!(!(a))", "{,}", "**/{a,*b}/!(c)"} {
		f.Add(seed, "a/b.c", uint8(0), uint8(0))
	}
	f.Fuzz(func(t *testing.T, pattern, s string, sep, flags uint8) {
		if len(pattern) > 256 || len(s) > 1024 {
			t.Skip()
		}
		opts := fuzzOptions(sep, flags)
		g, err := CompileWith(pattern, opts...)
		if err != nil {
			return
		}
		got := g.Match(s)
		_ = g.MatchBytes([]byte(s))
		again, err := CompileWith(g.String(), opts...)
		if err != nil {
			t.Fatalf("CompileWith(%q): expected no error, got: %v", g.String(), err)
		}
		if again.Match(s) != got {
			t.Fatalf("pattern %q, input %q: the glob of String does not agree", pattern, s)
		}
		ref, ok := reference(t, pattern, opts...)
		if !ok {
			t.Fatalf("pattern %q: CompileWith succeeded and the reference did not", pattern)
		}
		if want := ref.Match(s); got != want {
			t.Fatalf("pattern %q, input %q, separators %q, flags %d: expected %t, got: %t (matcher %s)", pattern, s, string(g.Separators()), flags, want, got, g.Matcher)
		}
	})
}
