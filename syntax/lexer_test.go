package syntax

import (
	"testing"
)

func TestLexGood(t *testing.T) {
	for id, test := range []struct {
		pattern string
		items   []Token
	}{
		{
			pattern: "",
			items: []Token{
				{TokenEOF, ""},
			},
		},
		{
			pattern: "hello",
			items: []Token{
				{TokenText, "hello"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "/{rate,[0-9]]}*",
			items: []Token{
				{TokenText, "/"},
				{TokenTermsOpen, "{"},
				{TokenText, "rate"},
				{TokenSeparator, ","},
				{TokenRangeOpen, "["},
				{TokenRangeLo, "0"},
				{TokenRangeBetween, "-"},
				{TokenRangeHi, "9"},
				{TokenRangeClose, "]"},
				{TokenText, "]"},
				{TokenTermsClose, "}"},
				{TokenAny, "*"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "hello,world",
			items: []Token{
				{TokenText, "hello,world"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "hello\\,world",
			items: []Token{
				{TokenText, "hello,world"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "hello\\{world",
			items: []Token{
				{TokenText, "hello{world"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "hello?",
			items: []Token{
				{TokenText, "hello"},
				{TokenSingle, "?"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "hellof*",
			items: []Token{
				{TokenText, "hellof"},
				{TokenAny, "*"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "hello**",
			items: []Token{
				{TokenText, "hello"},
				{TokenSuper, "**"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "[日-語]",
			items: []Token{
				{TokenRangeOpen, "["},
				{TokenRangeLo, "日"},
				{TokenRangeBetween, "-"},
				{TokenRangeHi, "語"},
				{TokenRangeClose, "]"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "[!日-語]",
			items: []Token{
				{TokenRangeOpen, "["},
				{TokenNot, "!"},
				{TokenRangeLo, "日"},
				{TokenRangeBetween, "-"},
				{TokenRangeHi, "語"},
				{TokenRangeClose, "]"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "[日本語]",
			items: []Token{
				{TokenRangeOpen, "["},
				{TokenText, "日本語"},
				{TokenRangeClose, "]"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "[!日本語]",
			items: []Token{
				{TokenRangeOpen, "["},
				{TokenNot, "!"},
				{TokenText, "日本語"},
				{TokenRangeClose, "]"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "{a,b}",
			items: []Token{
				{TokenTermsOpen, "{"},
				{TokenText, "a"},
				{TokenSeparator, ","},
				{TokenText, "b"},
				{TokenTermsClose, "}"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "/{z,ab}*",
			items: []Token{
				{TokenText, "/"},
				{TokenTermsOpen, "{"},
				{TokenText, "z"},
				{TokenSeparator, ","},
				{TokenText, "ab"},
				{TokenTermsClose, "}"},
				{TokenAny, "*"},
				{TokenEOF, ""},
			},
		},
		{
			pattern: "{[!日-語],*,?,{a,b,\\c}}",
			items: []Token{
				{TokenTermsOpen, "{"},
				{TokenRangeOpen, "["},
				{TokenNot, "!"},
				{TokenRangeLo, "日"},
				{TokenRangeBetween, "-"},
				{TokenRangeHi, "語"},
				{TokenRangeClose, "]"},
				{TokenSeparator, ","},
				{TokenAny, "*"},
				{TokenSeparator, ","},
				{TokenSingle, "?"},
				{TokenSeparator, ","},
				{TokenTermsOpen, "{"},
				{TokenText, "a"},
				{TokenSeparator, ","},
				{TokenText, "b"},
				{TokenSeparator, ","},
				{TokenText, "c"},
				{TokenTermsClose, "}"},
				{TokenTermsClose, "}"},
				{TokenEOF, ""},
			},
		},
	} {
		lexer := NewLexer(test.pattern)
		for i, exp := range test.items {
			act := lexer.Next()
			if act.Type != exp.Type {
				t.Errorf("#%d %q: wrong %d-th item type: exp: %q; act: %q\n\t(%s vs %s)", id, test.pattern, i, exp.Type, act.Type, exp, act)
			}
			if act.Raw != exp.Raw {
				t.Errorf("#%d %q: wrong %d-th item contents: exp: %q; act: %q\n\t(%s vs %s)", id, test.pattern, i, exp.Raw, act.Raw, exp, act)
			}
		}
	}
}
