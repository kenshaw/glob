package syntax

import (
	"reflect"
	"strconv"
	"testing"
)

func TestParse(t *testing.T) {
	for i, test := range []struct {
		tokens []Token
		exp    *Node
	}{
		{
			// pattern: "abc",
			tokens: []Token{
				{TokenText, "abc"},
				{TokenEOF, ""},
			},
			exp: New(Pattern, nil,
				New(Text, TextData{Text: "abc"}),
			),
		},
		{
			// pattern: "a*c",
			tokens: []Token{
				{TokenText, "a"},
				{TokenAny, "*"},
				{TokenText, "c"},
				{TokenEOF, ""},
			},
			exp: New(Pattern, nil,
				New(Text, TextData{Text: "a"}),
				New(Any, nil),
				New(Text, TextData{Text: "c"}),
			),
		},
		{
			// pattern: "a**c",
			tokens: []Token{
				{TokenText, "a"},
				{TokenSuper, "**"},
				{TokenText, "c"},
				{TokenEOF, ""},
			},
			exp: New(Pattern, nil,
				New(Text, TextData{Text: "a"}),
				New(Super, nil),
				New(Text, TextData{Text: "c"}),
			),
		},
		{
			// pattern: "a?c",
			tokens: []Token{
				{TokenText, "a"},
				{TokenSingle, "?"},
				{TokenText, "c"},
				{TokenEOF, ""},
			},
			exp: New(Pattern, nil,
				New(Text, TextData{Text: "a"}),
				New(Single, nil),
				New(Text, TextData{Text: "c"}),
			),
		},
		{
			// pattern: "[!a-z]",
			tokens: []Token{
				{TokenRangeOpen, "["},
				{TokenNot, "!"},
				{TokenRangeLo, "a"},
				{TokenRangeBetween, "-"},
				{TokenRangeHi, "z"},
				{TokenRangeClose, "]"},
				{TokenEOF, ""},
			},
			exp: New(Pattern, nil,
				New(Range, RangeData{Lo: 'a', Hi: 'z', Not: true}),
			),
		},
		{
			// pattern: "[az]",
			tokens: []Token{
				{TokenRangeOpen, "["},
				{TokenText, "az"},
				{TokenRangeClose, "]"},
				{TokenEOF, ""},
			},
			exp: New(Pattern, nil,
				New(List, ListData{Chars: "az"}),
			),
		},
		{
			// pattern: "{a,z}",
			tokens: []Token{
				{TokenTermsOpen, "{"},
				{TokenText, "a"},
				{TokenSeparator, ","},
				{TokenText, "z"},
				{TokenTermsClose, "}"},
				{TokenEOF, ""},
			},
			exp: New(Pattern, nil,
				New(AnyOf, nil,
					New(Pattern, nil,
						New(Text, TextData{Text: "a"}),
					),
					New(Pattern, nil,
						New(Text, TextData{Text: "z"}),
					),
				),
			),
		},
		{
			// pattern: "/{z,ab}*",
			tokens: []Token{
				{TokenText, "/"},
				{TokenTermsOpen, "{"},
				{TokenText, "z"},
				{TokenSeparator, ","},
				{TokenText, "ab"},
				{TokenTermsClose, "}"},
				{TokenAny, "*"},
				{TokenEOF, ""},
			},
			exp: New(Pattern, nil,
				New(Text, TextData{Text: "/"}),
				New(AnyOf, nil,
					New(Pattern, nil,
						New(Text, TextData{Text: "z"}),
					),
					New(Pattern, nil,
						New(Text, TextData{Text: "ab"}),
					),
				),
				New(Any, nil),
			),
		},
		{
			// pattern: "{a,{x,y},?,[a-z],[!qwe]}",
			tokens: []Token{
				{TokenTermsOpen, "{"},
				{TokenText, "a"},
				{TokenSeparator, ","},
				{TokenTermsOpen, "{"},
				{TokenText, "x"},
				{TokenSeparator, ","},
				{TokenText, "y"},
				{TokenTermsClose, "}"},
				{TokenSeparator, ","},
				{TokenSingle, "?"},
				{TokenSeparator, ","},
				{TokenRangeOpen, "["},
				{TokenRangeLo, "a"},
				{TokenRangeBetween, "-"},
				{TokenRangeHi, "z"},
				{TokenRangeClose, "]"},
				{TokenSeparator, ","},
				{TokenRangeOpen, "["},
				{TokenNot, "!"},
				{TokenText, "qwe"},
				{TokenRangeClose, "]"},
				{TokenTermsClose, "}"},
				{TokenEOF, ""},
			},
			exp: New(Pattern, nil,
				New(AnyOf, nil,
					New(Pattern, nil,
						New(Text, TextData{Text: "a"}),
					),
					New(Pattern, nil,
						New(AnyOf, nil,
							New(Pattern, nil,
								New(Text, TextData{Text: "x"}),
							),
							New(Pattern, nil,
								New(Text, TextData{Text: "y"}),
							),
						),
					),
					New(Pattern, nil,
						New(Single, nil),
					),
					New(Pattern, nil,
						New(Range, RangeData{Lo: 'a', Hi: 'z', Not: false}),
					),
					New(Pattern, nil,
						New(List, ListData{Chars: "qwe", Not: true}),
					),
				),
			),
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			l := &stubLexer{tokens: test.tokens}
			tree, err := parse(l)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(tree, test.exp) {
				t.Errorf("expected:\n%v\ngot:\n%v", test.exp, tree)
			}
		})
	}
}

type stubLexer struct {
	tokens []Token
	pos    int
}

func (s *stubLexer) Next() (ret Token) {
	if s.pos == len(s.tokens) {
		return Token{TokenEOF, ""}
	}
	ret = s.tokens[s.pos]
	s.pos++
	return
}
