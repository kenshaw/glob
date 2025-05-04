package syntax

import (
	"reflect"
	"testing"
)

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

func TestParseString(t *testing.T) {
	for id, test := range []struct {
		tokens []Token
		tree   *Node
	}{
		{
			// pattern: "abc",
			tokens: []Token{
				{TokenText, "abc"},
				{TokenEOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, Text{Text: "abc"}),
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
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, Text{Text: "a"}),
				NewNode(KindAny, nil),
				NewNode(KindText, Text{Text: "c"}),
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
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, Text{Text: "a"}),
				NewNode(KindSuper, nil),
				NewNode(KindText, Text{Text: "c"}),
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
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, Text{Text: "a"}),
				NewNode(KindSingle, nil),
				NewNode(KindText, Text{Text: "c"}),
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
			tree: NewNode(KindPattern, nil,
				NewNode(KindRange, Range{Lo: 'a', Hi: 'z', Not: true}),
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
			tree: NewNode(KindPattern, nil,
				NewNode(KindList, List{Chars: "az"}),
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
			tree: NewNode(KindPattern, nil,
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{Text: "a"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{Text: "z"}),
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
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, Text{Text: "/"}),
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{Text: "z"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{Text: "ab"}),
					),
				),
				NewNode(KindAny, nil),
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
			tree: NewNode(KindPattern, nil,
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{Text: "a"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindAnyOf, nil,
							NewNode(KindPattern, nil,
								NewNode(KindText, Text{Text: "x"}),
							),
							NewNode(KindPattern, nil,
								NewNode(KindText, Text{Text: "y"}),
							),
						),
					),
					NewNode(KindPattern, nil,
						NewNode(KindSingle, nil),
					),
					NewNode(KindPattern, nil,
						NewNode(KindRange, Range{Lo: 'a', Hi: 'z', Not: false}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindList, List{Chars: "qwe", Not: true}),
					),
				),
			),
		},
	} {
		lexer := &stubLexer{tokens: test.tokens}
		result, err := Parse(lexer)
		if err != nil {
			t.Errorf("[%d] unexpected error: %s", id, err)
		}
		if !reflect.DeepEqual(test.tree, result) {
			t.Errorf("[%d] Parse():\nact:\t%s\nexp:\t%s\n", id, result, test.tree)
		}
	}
}
