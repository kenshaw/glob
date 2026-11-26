package syntax

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

// TestLexerExtended holds the tokens of D9, D10 and D11.
func TestLexerExtended(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		pattern string
		items   []Token
	}{
		{"[a-cx]", []Token{
			{TokenRangeOpen, "["}, {TokenRangeLo, "a"}, {TokenRangeBetween, "-"}, {TokenRangeHi, "c"},
			{TokenText, "x"}, {TokenRangeClose, "]"}, {TokenEOF, ""},
		}},
		{"[^]a-]", []Token{
			{TokenRangeOpen, "["}, {TokenNot, "^"}, {TokenText, "]a-"}, {TokenRangeClose, "]"}, {TokenEOF, ""},
		}},
		{"[[:alpha:]_]", []Token{
			{TokenRangeOpen, "["}, {TokenClass, "alpha"}, {TokenText, "_"}, {TokenRangeClose, "]"}, {TokenEOF, ""},
		}},
		{"[[=a=][.-.]]", []Token{
			{TokenRangeOpen, "["}, {TokenText, "a-"}, {TokenRangeClose, "]"}, {TokenEOF, ""},
		}},
		{"[[:digit]", []Token{
			{TokenRangeOpen, "["}, {TokenText, "[:digit"}, {TokenRangeClose, "]"}, {TokenEOF, ""},
		}},
		{`[\]\-]`, []Token{
			{TokenRangeOpen, "["}, {TokenText, "]-"}, {TokenRangeClose, "]"}, {TokenEOF, ""},
		}},
		{"@(a|b)", []Token{
			{TokenExtOpen, "@("}, {TokenText, "a"}, {TokenExtSeparator, "|"}, {TokenText, "b"},
			{TokenExtClose, ")"}, {TokenEOF, ""},
		}},
		{"!(x)*", []Token{
			{TokenExtOpen, "!("}, {TokenText, "x"}, {TokenExtClose, ")"}, {TokenAny, "*"}, {TokenEOF, ""},
		}},
		{"**(a)", []Token{
			{TokenAny, "*"}, {TokenExtOpen, "*("}, {TokenText, "a"}, {TokenExtClose, ")"}, {TokenEOF, ""},
		}},
		{"a|b)c+d@e", []Token{
			{TokenText, "a|b)c+d@e"}, {TokenEOF, ""},
		}},
		{"{a,+(b|c)}", []Token{
			{TokenTermsOpen, "{"}, {TokenText, "a"}, {TokenSeparator, ","}, {TokenExtOpen, "+("},
			{TokenText, "b"}, {TokenExtSeparator, "|"}, {TokenText, "c"}, {TokenExtClose, ")"},
			{TokenTermsClose, "}"}, {TokenEOF, ""},
		}},
		{"@(a,b)", []Token{
			{TokenExtOpen, "@("}, {TokenText, "a,b"}, {TokenExtClose, ")"}, {TokenEOF, ""},
		}},
		{"x{01..10..2}y", []Token{
			{TokenText, "x"}, {TokenSequence, "01..10..2"}, {TokenText, "y"}, {TokenEOF, ""},
		}},
		{"{1..a}", []Token{
			{TokenTermsOpen, "{"}, {TokenText, "1..a"}, {TokenTermsClose, "}"}, {TokenEOF, ""},
		}},
		{"a\x00b", []Token{
			{TokenText, "a\x00b"}, {TokenEOF, ""},
		}},
		{"{a", []Token{
			{TokenError, `unexpected end of input: unclosed "{"`},
		}},
		{`a\`, []Token{
			{TokenError, `unexpected end of input after "\\"`},
		}},
		{"[[:foo:]]", []Token{
			{TokenError, `unknown character class "foo"`},
		}},
	} {
		l := NewLexer(test.pattern)
		var got []Token
		for {
			token := l.Next()
			got = append(got, token)
			if token.Token == TokenEOF || token.Token == TokenError {
				break
			}
		}
		if !slices.Equal(got, test.items) {
			t.Errorf("%q: expected %v, got: %v", test.pattern, test.items, got)
		}
	}
}

// TestParseExtended holds the nodes of D9, D10 and D11.
func TestParseExtended(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		pattern string
		exp     *Node
	}{
		{"[a-cx]", New(Pattern, nil, New(Class, ClassData{Chars: "x", Ranges: []RangeData{{Lo: 'a', Hi: 'c'}}}))},
		{"[!a-c]", New(Pattern, nil, New(Range, RangeData{Lo: 'a', Hi: 'c', Not: true}))},
		{"[^ab]", New(Pattern, nil, New(List, ListData{Chars: "ab", Not: true}))},
		{"[[:digit:]]", New(Pattern, nil, New(Class, ClassData{Classes: []string{"digit"}}))},
		{"@(a|b)", New(Pattern, nil, New(Ext, ExtData{Op: '@'},
			New(Pattern, nil, New(Text, TextData{"a"})),
			New(Pattern, nil, New(Text, TextData{"b"})),
		))},
		{"!()", New(Pattern, nil, New(Ext, ExtData{Op: '!'}, New(Pattern, nil)))},
		{"{1..3}", New(Pattern, nil, New(AnyOf, nil,
			New(Pattern, nil, New(Text, TextData{"1"})),
			New(Pattern, nil, New(Text, TextData{"2"})),
			New(Pattern, nil, New(Text, TextData{"3"})),
		))},
	} {
		got, err := Parse(NewLexer(test.pattern))
		if err != nil {
			t.Errorf("%q: expected no error, got: %v", test.pattern, err)
			continue
		}
		if !got.Equal(test.exp) {
			t.Errorf("%q: expected %s, got: %s", test.pattern, test.exp, got)
		}
	}
}

// TestExpandSequence holds D11.
func TestExpandSequence(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		raw string
		exp string
	}{
		{"1..5", "1 2 3 4 5"},
		{"5..1", "5 4 3 2 1"},
		{"-2..2", "-2 -1 0 1 2"},
		{"+1..3", "1 2 3"},
		{"01..10..3", "01 04 07 10"},
		{"1..010..4", "001 005 009"},
		{"-05..5..5", "-05 000 005"},
		{"10..1..-3", "10 7 4 1"},
		{"1..3..0", "1 2 3"},
		{"0..20..05", "0 5 10 15 20"},
		{"a..e", "a b c d e"},
		{"e..a..2", "e c a"},
		{"Y..b", "Y Z [ \\ ] ^ _ ` a b"},
		{"a..z..1000", "a"},
	} {
		n, err := expandSequence(test.raw)
		if err != nil {
			t.Errorf("%q: expected no error, got: %v", test.raw, err)
			continue
		}
		var items []string
		for _, c := range n.Children {
			items = append(items, c.Children[0].Value.(TextData).Text)
		}
		if got := strings.Join(items, " "); got != test.exp {
			t.Errorf("%q: expected %q, got: %q", test.raw, test.exp, got)
		}
	}
	for _, raw := range []string{"0..10000", "-5000..5000"} {
		if _, err := expandSequence(raw); err == nil {
			t.Errorf("%q: expected an error, got: nil", raw)
		}
	}
	if _, err := expandSequence("1..10000"); err != nil {
		t.Errorf("expected no error for %d items, got: %v", MaxSequence, err)
	}
}

// TestClassData holds D9 and D17.
func TestClassData(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		data ClassData
		in   string
		out  string
	}{
		{ClassData{Chars: "abc"}, "abc", "dA-"},
		{ClassData{Not: true, Chars: "abc"}, "dA-", "abc"},
		{ClassData{Ranges: []RangeData{{Lo: 'a', Hi: 'c'}, {Lo: '0', Hi: '2'}}}, "ab012", "d3A"},
		{ClassData{Classes: []string{"alpha"}}, "aZéж", "0_ !"},
		{ClassData{Classes: []string{"digit"}}, "0189", "a٣"},
		{ClassData{Classes: []string{"xdigit"}}, "09afAF", "gG"},
		{ClassData{Classes: []string{"alnum"}}, "a9é", "_-"},
		{ClassData{Classes: []string{"word"}}, "a9_é", "-."},
		{ClassData{Classes: []string{"upper"}}, "AÉ", "aé1"},
		{ClassData{Classes: []string{"lower"}}, "aé", "AÉ1"},
		{ClassData{Classes: []string{"space"}}, " \t\n\v\f\r ", "a_"},
		{ClassData{Classes: []string{"blank"}}, " \t ", "\na"},
		{ClassData{Classes: []string{"punct"}}, "!-_~€", "a0 "},
		{ClassData{Classes: []string{"cntrl"}}, "\x00\x1f\x7f", "a "},
		{ClassData{Classes: []string{"print"}}, "a ~é", "\x00\t"},
		{ClassData{Classes: []string{"graph"}}, "a~é", " \t\x00"},
		{ClassData{Classes: []string{"ascii"}}, "a\x00\x7f", "é\u0080"},
		{ClassData{Fold: true, Ranges: []RangeData{{Lo: 'a', Hi: 'c'}}}, "abcABC", "dD"},
		{ClassData{Fold: true, Chars: "k"}, "kKK", "j"},
		{ClassData{Fold: true, Not: true, Chars: "k"}, "j", "kKK"},
		{ClassData{Fold: true, Classes: []string{"upper"}}, "aA", "1"},
	} {
		for _, r := range test.in {
			if !test.data.Contains(r) {
				t.Errorf("%s: expected %q in the class", test.data, r)
			}
		}
		for _, r := range test.out {
			if test.data.Contains(r) {
				t.Errorf("%s: expected %q out of the class", test.data, r)
			}
		}
	}
	for name := range classes {
		if !isClassName(name) {
			t.Errorf("expected %q to be a class name", name)
		}
	}
	if len(classes) != 14 {
		t.Errorf("expected 14 named classes, got: %d", len(classes))
	}
}

func TestClassMatcher(t *testing.T) {
	t.Parallel()
	m := NewClass(ClassData{Chars: "x", Ranges: []RangeData{{Lo: 'a', Hi: 'c'}}})
	for s, exp := range map[string]bool{"a": true, "x": true, "": false, "ab": false, "d": false} {
		if got := m.Match(s); got != exp {
			t.Errorf("%q: expected %t, got: %t", s, exp, got)
		}
	}
	i, seg := m.Index("zzxa")
	if i != 2 || !reflect.DeepEqual(seg, []int{1}) {
		t.Errorf("expected 2 [1], got: %d %v", i, seg)
	}
	if i, _ := m.Index("zz"); i != -1 {
		t.Errorf("expected -1, got: %d", i)
	}
	if m.Len() != 1 || m.Size() != 1 {
		t.Errorf("expected a length and a size of 1, got: %d and %d", m.Len(), m.Size())
	}
}

// TestFold holds D17.
func TestFold(t *testing.T) {
	t.Parallel()
	for a, b := range map[string]string{
		"hello": "HELLO",
		"k":     "K",
		"s":     "ſ",
		"ǅ":     "ǆ",
		"ß":     "ẞ",
	} {
		if FoldString(a) != FoldString(b) {
			t.Errorf("expected %q and %q to fold to one string, got: %q and %q", a, b, FoldString(a), FoldString(b))
		}
	}
	if FoldString("straße") == FoldString("STRASSE") {
		t.Errorf("expected straße and STRASSE to differ")
	}
	tree, err := Parse(NewLexer("Ab[c-d][!E]"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	folded := Fold(tree)
	if folded.Children[0].Value.(TextData).Text != FoldString("ab") {
		t.Errorf("expected the text to fold, got: %s", folded)
	}
	for _, c := range folded.Children[1:] {
		if c.Type != Class || !c.Value.(ClassData).Fold {
			t.Errorf("expected a class that folds, got: %s", c)
		}
	}
	if tree.Children[1].Type != Range {
		t.Errorf("Fold changed the tree it copies: %s", tree)
	}
}

// TestGlobstar holds D16.
func TestGlobstar(t *testing.T) {
	t.Parallel()
	sep := []rune{'/'}
	for _, test := range []struct {
		pattern string
		changed bool
	}{
		{"**/a", true},
		{"a/**/b", true},
		{"a/***/b", true},
		{"{x/,}**/b", true},
		{"a**/b", false},
		{"a/**", false},
		{"a/**b", false},
		{"{a,b}**/c", false},
		{"a/*/b", false},
	} {
		tree, err := Parse(NewLexer(test.pattern))
		if err != nil {
			t.Fatalf("%q: expected no error, got: %v", test.pattern, err)
		}
		got := Globstar(tree, sep)
		if changed := !got.Equal(tree); changed != test.changed {
			t.Errorf("%q: expected changed %t, got: %t (%s)", test.pattern, test.changed, changed, got)
		}
	}
}

// TestBacktrack holds D10 and D12.
func TestBacktrack(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		pattern, s string
		sep        []rune
		exp        bool
	}{
		{"@(a|b)c", "bc", nil, true},
		{"@(a|b)c", "abc", nil, false},
		{"?(a|b)c", "c", nil, true},
		{"?(a|b)c", "abc", nil, false},
		{"*(ab)", "", nil, true},
		{"*(ab)", "ababab", nil, true},
		{"*(ab)", "aba", nil, false},
		{"+(ab)", "", nil, false},
		{"+(ab|c)", "abcab", nil, true},
		{"+(|a)", "", nil, true},
		{"!(a)", "", nil, true},
		{"!(a)", "a", nil, false},
		{"!(a)", "b", nil, true},
		{"!(a)*", "a", nil, true},
		{"!(!(a))", "a", nil, true},
		{"!(!(a))", "b", nil, false},
		{"!(*.go)", "x.go", nil, false},
		{"!(*.go)", "x.c", nil, true},
		{"!(x)", "a/b", []rune{'/'}, false},
		{"?(x/)y", "x/y", []rune{'/'}, true},
		{"a*!(z)", "az", nil, true},
		{"@(a|b)*", "a/b", []rune{'/'}, false},
		{"@(a|b)**", "a/b", []rune{'/'}, true},
		{"+(a|b|ab)+(c)", "abababcc", nil, true},
		{"[[:digit:]]+([0-9])", "2026", nil, true},
		{"{a,@(b|c)}d", "cd", nil, true},
		{"x?y", "x/y", []rune{'/'}, false},
		{"x[/]y", "x/y", []rune{'/'}, true},
	} {
		tree, err := Parse(NewLexer(test.pattern))
		if err != nil {
			t.Fatalf("%q: expected no error, got: %v", test.pattern, err)
		}
		m := NewBacktrack(tree, test.sep)
		if got := m.Match(test.s); got != test.exp {
			t.Errorf("%q, %q: expected %t, got: %t", test.pattern, test.s, test.exp, got)
		}
		if m.Len() != 0 || !strings.HasPrefix(m.String(), "<backtrack:") {
			t.Errorf("%q: unexpected length %d or string %s", test.pattern, m.Len(), m)
		}
	}
}

// TestBacktrackLongInput makes sure that the matcher remembers its answers.
// Without them, each pattern takes exponential time in the length of the
// input.
func TestBacktrackLongInput(t *testing.T) {
	t.Parallel()
	s := strings.Repeat("a", 200)
	for _, pattern := range []string{"*(a|aa)b", "+(a|aa)b", "*a*a*a*a*a*b", "!(b)!(b)!(b)b"} {
		tree, err := Parse(NewLexer(pattern))
		if err != nil {
			t.Fatalf("%q: expected no error, got: %v", pattern, err)
		}
		if NewBacktrack(tree, nil).Match(s) {
			t.Errorf("%q: expected no match", pattern)
		}
	}
}
