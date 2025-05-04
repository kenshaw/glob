package syntax

import (
	"reflect"
	"testing"

	"github.com/kenshaw/glob/match"
)

var separators = []rune{'.'}

func TestCompiler(t *testing.T) {
	for _, test := range []struct {
		name string
		ast  *Node
		exp  match.Matcher
		sep  []rune
	}{
		{
			// #0
			ast: NewNode(KindPattern, nil,
				NewNode(KindText, Text{"abc"}),
			),
			exp: match.NewText("abc"),
		},
		{
			// #1
			ast: NewNode(KindPattern, nil,
				NewNode(KindAny, nil),
			),
			sep: separators,
			exp: match.NewAny(separators),
		},
		{
			// #2
			ast: NewNode(KindPattern, nil,
				NewNode(KindAny, nil),
			),
			exp: match.NewSuper(),
		},
		{
			// #3
			ast: NewNode(KindPattern, nil,
				NewNode(KindSuper, nil),
			),
			exp: match.NewSuper(),
		},
		{
			// #4
			ast: NewNode(KindPattern, nil,
				NewNode(KindSingle, nil),
			),
			sep: separators,
			exp: match.NewSingle(separators),
		},
		{
			// #5
			ast: NewNode(KindPattern, nil,
				NewNode(KindRange, Range{
					Lo:  'a',
					Hi:  'z',
					Not: true,
				}),
			),
			exp: match.NewRange('a', 'z', true),
		},
		{
			// #6
			ast: NewNode(KindPattern, nil,
				NewNode(KindList, List{
					Chars: "abc",
					Not:   true,
				}),
			),
			exp: match.NewList([]rune{'a', 'b', 'c'}, true),
		},
		{
			// #7
			ast: NewNode(KindPattern, nil,
				NewNode(KindAny, nil),
				NewNode(KindSingle, nil),
				NewNode(KindSingle, nil),
				NewNode(KindSingle, nil),
			),
			sep: separators,
			exp: match.NewEveryOf([]match.Matcher{
				match.NewMin(3),
				match.NewAny(separators),
			}),
		},
		{
			// #8
			ast: NewNode(KindPattern, nil,
				NewNode(KindAny, nil),
				NewNode(KindSingle, nil),
				NewNode(KindSingle, nil),
				NewNode(KindSingle, nil),
			),
			exp: match.NewMin(3),
		},
		{
			// #9
			ast: NewNode(KindPattern, nil,
				NewNode(KindAny, nil),
				NewNode(KindText, Text{"abc"}),
				NewNode(KindSingle, nil),
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
			ast: NewNode(KindPattern, nil,
				NewNode(KindText, Text{"/"}),
				NewNode(KindAnyOf, nil,
					NewNode(KindText, Text{"z"}),
					NewNode(KindText, Text{"ab"}),
				),
				NewNode(KindSuper, nil),
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
			ast: NewNode(KindPattern, nil,
				NewNode(KindSuper, nil),
				NewNode(KindSingle, nil),
				NewNode(KindText, Text{"abc"}),
				NewNode(KindSingle, nil),
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
			ast: NewNode(KindPattern, nil,
				NewNode(KindAny, nil),
				NewNode(KindText, Text{"abc"}),
			),
			exp: match.NewSuffix("abc"),
		},
		{
			// #13
			ast: NewNode(KindPattern, nil,
				NewNode(KindText, Text{"abc"}),
				NewNode(KindAny, nil),
			),
			exp: match.NewPrefix("abc"),
		},
		{
			// #14
			ast: NewNode(KindPattern, nil,
				NewNode(KindText, Text{"abc"}),
				NewNode(KindAny, nil),
				NewNode(KindText, Text{"def"}),
			),
			exp: match.NewPrefixSuffix("abc", "def"),
		},
		{
			// #15
			ast: NewNode(KindPattern, nil,
				NewNode(KindAny, nil),
				NewNode(KindAny, nil),
				NewNode(KindAny, nil),
				NewNode(KindText, Text{"abc"}),
				NewNode(KindAny, nil),
				NewNode(KindAny, nil),
			),
			exp: match.NewContains("abc"),
		},
		{
			// #16
			ast: NewNode(KindPattern, nil,
				NewNode(KindAny, nil),
				NewNode(KindAny, nil),
				NewNode(KindAny, nil),
				NewNode(KindText, Text{"abc"}),
				NewNode(KindAny, nil),
				NewNode(KindAny, nil),
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
			ast: NewNode(KindPattern, nil,
				NewNode(KindSuper, nil),
				NewNode(KindSingle, nil),
				NewNode(KindText, Text{"abc"}),
				NewNode(KindSuper, nil),
				NewNode(KindSingle, nil),
			),
			exp: match.NewTree(
				match.NewText("abc"),
				match.NewMin(1),
				match.NewMin(1),
			),
		},
		{
			// #18
			ast: NewNode(KindPattern, nil,
				NewNode(KindText, Text{"abc"}),
			),
			exp: match.NewText("abc"),
		},
		{
			// #19
			ast: NewNode(KindPattern, nil,
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindAnyOf, nil,
							NewNode(KindPattern, nil,
								NewNode(KindText, Text{"abc"}),
							),
						),
					),
				),
			),
			exp: match.NewText("abc"),
		},
		{
			// #20
			ast: NewNode(KindPattern, nil,
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{"abc"}),
						NewNode(KindSingle, nil),
					),
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{"abc"}),
						NewNode(KindList, List{Chars: "def"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{"abc"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{"abc"}),
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
			ast: NewNode(KindPattern, nil,
				NewNode(KindRange, Range{Lo: 'a', Hi: 'z'}),
				NewNode(KindRange, Range{Lo: 'a', Hi: 'x', Not: true}),
				NewNode(KindAny, nil),
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
			ast: NewNode(KindPattern, nil,
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{"abc"}),
						NewNode(KindList, List{Chars: "abc"}),
						NewNode(KindText, Text{"ghi"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{"abc"}),
						NewNode(KindList, List{Chars: "def"}),
						NewNode(KindText, Text{"ghi"}),
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
		t.Run(test.name, func(t *testing.T) {
			act, err := test.ast.Compile(test.sep)
			if err != nil {
				t.Fatalf("compilation error: %s", err)
			}
			if !reflect.DeepEqual(act, test.exp) {
				t.Errorf(
					"Compile():\nact: %#v\nexp: %#v\n\ngraphviz:\n%s\n%s\n",
					act, test.exp,
					match.Graphviz("act", act.(match.Matcher)),
					match.Graphviz("exp", test.exp.(match.Matcher)),
				)
			}
		})
	}
}
