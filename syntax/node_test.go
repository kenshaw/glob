package syntax

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/kenshaw/glob/match"
)

func TestNode(t *testing.T) {
	for i, test := range []struct {
		tree *Node
		exp  match.Matcher
		sep  []rune
	}{
		{
			// #0
			tree: New(KindPattern, nil,
				New(KindText, Text{"abc"}),
			),
			exp: match.NewText("abc"),
		},
		{
			// #1
			tree: New(KindPattern, nil,
				New(KindAny, nil),
			),
			sep: separators,
			exp: match.NewAny(separators),
		},
		{
			// #2
			tree: New(KindPattern, nil,
				New(KindAny, nil),
			),
			exp: match.NewSuper(),
		},
		{
			// #3
			tree: New(KindPattern, nil,
				New(KindSuper, nil),
			),
			exp: match.NewSuper(),
		},
		{
			// #4
			tree: New(KindPattern, nil,
				New(KindSingle, nil),
			),
			sep: separators,
			exp: match.NewSingle(separators),
		},
		{
			// #5
			tree: New(KindPattern, nil,
				New(KindRange, Range{
					Lo:  'a',
					Hi:  'z',
					Not: true,
				}),
			),
			exp: match.NewRange('a', 'z', true),
		},
		{
			// #6
			tree: New(KindPattern, nil,
				New(KindList, List{
					Chars: "abc",
					Not:   true,
				}),
			),
			exp: match.NewList([]rune{'a', 'b', 'c'}, true),
		},
		{
			// #7
			tree: New(KindPattern, nil,
				New(KindAny, nil),
				New(KindSingle, nil),
				New(KindSingle, nil),
				New(KindSingle, nil),
			),
			sep: separators,
			exp: match.NewEveryOf([]match.Matcher{
				match.NewMin(3),
				match.NewAny(separators),
			}),
		},
		{
			// #8
			tree: New(KindPattern, nil,
				New(KindAny, nil),
				New(KindSingle, nil),
				New(KindSingle, nil),
				New(KindSingle, nil),
			),
			exp: match.NewMin(3),
		},
		{
			// #9
			tree: New(KindPattern, nil,
				New(KindAny, nil),
				New(KindText, Text{"abc"}),
				New(KindSingle, nil),
			),
			sep: separators,
			exp: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewText("abc"),
					match.NewSingle(separators),
				}),
				match.NewAny(separators),
				match.Nothing{},
			),
		},
		{
			// #10
			tree: New(KindPattern, nil,
				New(KindText, Text{"/"}),
				New(KindAnyOf, nil,
					New(KindText, Text{"z"}),
					New(KindText, Text{"ab"}),
				),
				New(KindSuper, nil),
			),
			sep: separators,
			exp: match.NewTree(
				match.NewText("/"),
				match.Nothing{},
				match.NewTree(
					match.MustIndexedAnyOf(
						match.NewText("z"),
						match.NewText("ab"),
					),
					match.Nothing{},
					match.NewSuper(),
				),
			),
		},
		{
			// #11
			tree: New(KindPattern, nil,
				New(KindSuper, nil),
				New(KindSingle, nil),
				New(KindText, Text{"abc"}),
				New(KindSingle, nil),
			),
			sep: separators,
			exp: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewSingle(separators),
					match.NewText("abc"),
					match.NewSingle(separators),
				}),
				match.NewSuper(),
				match.Nothing{},
			),
		},
		{
			// #12
			tree: New(KindPattern, nil,
				New(KindAny, nil),
				New(KindText, Text{"abc"}),
			),
			exp: match.NewSuffix("abc"),
		},
		{
			// #13
			tree: New(KindPattern, nil,
				New(KindText, Text{"abc"}),
				New(KindAny, nil),
			),
			exp: match.NewPrefix("abc"),
		},
		{
			// #14
			tree: New(KindPattern, nil,
				New(KindText, Text{"abc"}),
				New(KindAny, nil),
				New(KindText, Text{"def"}),
			),
			exp: match.NewPrefixSuffix("abc", "def"),
		},
		{
			// #15
			tree: New(KindPattern, nil,
				New(KindAny, nil),
				New(KindAny, nil),
				New(KindAny, nil),
				New(KindText, Text{"abc"}),
				New(KindAny, nil),
				New(KindAny, nil),
			),
			exp: match.NewContains("abc"),
		},
		{
			// #16
			tree: New(KindPattern, nil,
				New(KindAny, nil),
				New(KindAny, nil),
				New(KindAny, nil),
				New(KindText, Text{"abc"}),
				New(KindAny, nil),
				New(KindAny, nil),
			),
			sep: separators,
			exp: match.NewTree(
				match.NewText("abc"),
				match.NewAny(separators),
				match.NewAny(separators),
			),
		},
		{
			// #17
			// pattern: "**?abc**?"
			tree: New(KindPattern, nil,
				New(KindSuper, nil),
				New(KindSingle, nil),
				New(KindText, Text{"abc"}),
				New(KindSuper, nil),
				New(KindSingle, nil),
			),
			exp: match.NewTree(
				match.NewText("abc"),
				match.NewMin(1),
				match.NewMin(1),
			),
		},
		{
			// #18
			tree: New(KindPattern, nil,
				New(KindText, Text{"abc"}),
			),
			exp: match.NewText("abc"),
		},
		{
			// #19
			tree: New(KindPattern, nil,
				New(KindAnyOf, nil,
					New(KindPattern, nil,
						New(KindAnyOf, nil,
							New(KindPattern, nil,
								New(KindText, Text{"abc"}),
							),
						),
					),
				),
			),
			exp: match.NewText("abc"),
		},
		{
			// #20
			tree: New(KindPattern, nil,
				New(KindAnyOf, nil,
					New(KindPattern, nil,
						New(KindText, Text{"abc"}),
						New(KindSingle, nil),
					),
					New(KindPattern, nil,
						New(KindText, Text{"abc"}),
						New(KindList, List{Chars: "def"}),
					),
					New(KindPattern, nil,
						New(KindText, Text{"abc"}),
					),
					New(KindPattern, nil,
						New(KindText, Text{"abc"}),
					),
				),
			),
			exp: match.NewTree(
				match.NewText("abc"),
				match.Nothing{},
				match.NewAnyOf(
					match.NewSingle(nil),
					match.NewList([]rune{'d', 'e', 'f'}, false),
					match.NewNothing(),
				),
			),
		},
		{
			// #21
			tree: New(KindPattern, nil,
				New(KindRange, Range{Lo: 'a', Hi: 'z'}),
				New(KindRange, Range{Lo: 'a', Hi: 'x', Not: true}),
				New(KindAny, nil),
			),
			exp: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewRange('a', 'z', false),
					match.NewRange('a', 'x', true),
				}),
				match.Nothing{},
				match.NewSuper(),
			),
		},
		{
			// #22
			tree: New(KindPattern, nil,
				New(KindAnyOf, nil,
					New(KindPattern, nil,
						New(KindText, Text{"abc"}),
						New(KindList, List{Chars: "abc"}),
						New(KindText, Text{"ghi"}),
					),
					New(KindPattern, nil,
						New(KindText, Text{"abc"}),
						New(KindList, List{Chars: "def"}),
						New(KindText, Text{"ghi"}),
					),
				),
			),
			exp: match.NewRow([]match.MatchIndexSizer{
				match.NewText("abc"),
				match.MustIndexedSizedAnyOf(
					match.NewList([]rune{'a', 'b', 'c'}, false),
					match.NewList([]rune{'d', 'e', 'f'}, false),
				),
				match.NewText("ghi"),
			}),
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			m, err := test.tree.Compile(test.sep)
			if err != nil {
				t.Fatalf("compilation error: %s", err)
			}
			if !reflect.DeepEqual(m, test.exp) {
				t.Errorf(
					"Compile():\nact: %#v\nexp: %#v\n\ngraphviz:\n%s\n%s\n",
					m, test.exp,
					match.Graphviz("act", m.(match.Matcher)),
					match.Graphviz("exp", test.exp.(match.Matcher)),
				)
			}
		})
	}
}

var separators = []rune{'.'}
