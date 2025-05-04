package glob

import (
	"reflect"
	"testing"

	"github.com/kenshaw/glob/match"
	"github.com/kenshaw/glob/syntax"
)

var separators = []rune{'.'}

func TestCompiler(t *testing.T) {
	for _, test := range []struct {
		name string
		ast  *syntax.Node
		exp  match.Matcher
		sep  []rune
	}{
		{
			// #0
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
			),
			exp: match.NewText("abc"),
		},
		{
			// #1
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindAny, nil),
			),
			sep: separators,
			exp: match.NewAny(separators),
		},
		{
			// #2
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindAny, nil),
			),
			exp: match.NewSuper(),
		},
		{
			// #3
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindSuper, nil),
			),
			exp: match.NewSuper(),
		},
		{
			// #4
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindSingle, nil),
			),
			sep: separators,
			exp: match.NewSingle(separators),
		},
		{
			// #5
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindRange, syntax.Range{
					Lo:  'a',
					Hi:  'z',
					Not: true,
				}),
			),
			exp: match.NewRange('a', 'z', true),
		},
		{
			// #6
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindList, syntax.List{
					Chars: "abc",
					Not:   true,
				}),
			),
			exp: match.NewList([]rune{'a', 'b', 'c'}, true),
		},
		{
			// #7
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindSingle, nil),
				syntax.NewNode(syntax.KindSingle, nil),
				syntax.NewNode(syntax.KindSingle, nil),
			),
			sep: separators,
			exp: match.NewEveryOf([]match.Matcher{
				match.NewMin(3),
				match.NewAny(separators),
			}),
		},
		{
			// #8
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindSingle, nil),
				syntax.NewNode(syntax.KindSingle, nil),
				syntax.NewNode(syntax.KindSingle, nil),
			),
			exp: match.NewMin(3),
		},
		{
			// #9
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
				syntax.NewNode(syntax.KindSingle, nil),
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
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindText, syntax.Text{"/"}),
				syntax.NewNode(syntax.KindAnyOf, nil,
					syntax.NewNode(syntax.KindText, syntax.Text{"z"}),
					syntax.NewNode(syntax.KindText, syntax.Text{"ab"}),
				),
				syntax.NewNode(syntax.KindSuper, nil),
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
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindSuper, nil),
				syntax.NewNode(syntax.KindSingle, nil),
				syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
				syntax.NewNode(syntax.KindSingle, nil),
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
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
			),
			exp: match.NewSuffix("abc"),
		},
		{
			// #13
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
				syntax.NewNode(syntax.KindAny, nil),
			),
			exp: match.NewPrefix("abc"),
		},
		{
			// #14
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindText, syntax.Text{"def"}),
			),
			exp: match.NewPrefixSuffix("abc", "def"),
		},
		{
			// #15
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindAny, nil),
			),
			exp: match.NewContains("abc"),
		},
		{
			// #16
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
				syntax.NewNode(syntax.KindAny, nil),
				syntax.NewNode(syntax.KindAny, nil),
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
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindSuper, nil),
				syntax.NewNode(syntax.KindSingle, nil),
				syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
				syntax.NewNode(syntax.KindSuper, nil),
				syntax.NewNode(syntax.KindSingle, nil),
			),
			exp: match.NewTree(
				match.NewText("abc"),
				match.NewMin(1),
				match.NewMin(1),
			),
		},
		{
			// #18
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
			),
			exp: match.NewText("abc"),
		},
		{
			// #19
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindAnyOf, nil,
					syntax.NewNode(syntax.KindPattern, nil,
						syntax.NewNode(syntax.KindAnyOf, nil,
							syntax.NewNode(syntax.KindPattern, nil,
								syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
							),
						),
					),
				),
			),
			exp: match.NewText("abc"),
		},
		{
			// #20
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindAnyOf, nil,
					syntax.NewNode(syntax.KindPattern, nil,
						syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
						syntax.NewNode(syntax.KindSingle, nil),
					),
					syntax.NewNode(syntax.KindPattern, nil,
						syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
						syntax.NewNode(syntax.KindList, syntax.List{Chars: "def"}),
					),
					syntax.NewNode(syntax.KindPattern, nil,
						syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
					),
					syntax.NewNode(syntax.KindPattern, nil,
						syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
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
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindRange, syntax.Range{Lo: 'a', Hi: 'z'}),
				syntax.NewNode(syntax.KindRange, syntax.Range{Lo: 'a', Hi: 'x', Not: true}),
				syntax.NewNode(syntax.KindAny, nil),
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
			ast: syntax.NewNode(syntax.KindPattern, nil,
				syntax.NewNode(syntax.KindAnyOf, nil,
					syntax.NewNode(syntax.KindPattern, nil,
						syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
						syntax.NewNode(syntax.KindList, syntax.List{Chars: "abc"}),
						syntax.NewNode(syntax.KindText, syntax.Text{"ghi"}),
					),
					syntax.NewNode(syntax.KindPattern, nil,
						syntax.NewNode(syntax.KindText, syntax.Text{"abc"}),
						syntax.NewNode(syntax.KindList, syntax.List{Chars: "def"}),
						syntax.NewNode(syntax.KindText, syntax.Text{"ghi"}),
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
			act, err := CompileTree(test.ast, test.sep)
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
