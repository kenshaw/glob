package syntax

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/kenshaw/glob/match"
)

func TestNode(t *testing.T) {
	sep := []rune{'.'}
	for i, test := range []struct {
		tree *Node
		sep  []rune
		exp  match.Matcher
	}{
		{
			// #0
			tree: New(Pattern, nil,
				New(Text, TextData{"abc"}),
			),
			exp: match.NewText("abc"),
		},
		{
			// #1
			tree: New(Pattern, nil,
				New(Any, nil),
			),
			sep: sep,
			exp: match.NewAny(sep),
		},
		{
			// #2
			tree: New(Pattern, nil,
				New(Any, nil),
			),
			exp: match.NewSuper(),
		},
		{
			// #3
			tree: New(Pattern, nil,
				New(Super, nil),
			),
			exp: match.NewSuper(),
		},
		{
			// #4
			tree: New(Pattern, nil,
				New(Single, nil),
			),
			sep: sep,
			exp: match.NewSingle(sep),
		},
		{
			// #5
			tree: New(Pattern, nil,
				New(Range, RangeData{
					Lo:  'a',
					Hi:  'z',
					Not: true,
				}),
			),
			exp: match.NewRange('a', 'z', true),
		},
		{
			// #6
			tree: New(Pattern, nil,
				New(List, ListData{
					Chars: "abc",
					Not:   true,
				}),
			),
			exp: match.NewList([]rune{'a', 'b', 'c'}, true),
		},
		{
			// #7
			tree: New(Pattern, nil,
				New(Any, nil),
				New(Single, nil),
				New(Single, nil),
				New(Single, nil),
			),
			sep: sep,
			exp: match.NewEveryOf([]match.Matcher{
				match.NewMin(3),
				match.NewAny(sep),
			}),
		},
		{
			// #8
			tree: New(Pattern, nil,
				New(Any, nil),
				New(Single, nil),
				New(Single, nil),
				New(Single, nil),
			),
			exp: match.NewMin(3),
		},
		{
			// #9
			tree: New(Pattern, nil,
				New(Any, nil),
				New(Text, TextData{"abc"}),
				New(Single, nil),
			),
			sep: sep,
			exp: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewText("abc"),
					match.NewSingle(sep),
				}),
				match.NewAny(sep),
				match.NothingMatcher{},
			),
		},
		{
			// #10
			tree: New(Pattern, nil,
				New(Text, TextData{"/"}),
				New(AnyOf, nil,
					New(Text, TextData{"z"}),
					New(Text, TextData{"ab"}),
				),
				New(Super, nil),
			),
			sep: sep,
			exp: match.NewTree(
				match.NewText("/"),
				match.NothingMatcher{},
				match.NewTree(
					match.MustIndexedAnyOf(
						match.NewText("z"),
						match.NewText("ab"),
					),
					match.NothingMatcher{},
					match.NewSuper(),
				),
			),
		},
		{
			// #11
			tree: New(Pattern, nil,
				New(Super, nil),
				New(Single, nil),
				New(Text, TextData{"abc"}),
				New(Single, nil),
			),
			sep: sep,
			exp: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewSingle(sep),
					match.NewText("abc"),
					match.NewSingle(sep),
				}),
				match.NewSuper(),
				match.NothingMatcher{},
			),
		},
		{
			// #12
			tree: New(Pattern, nil,
				New(Any, nil),
				New(Text, TextData{"abc"}),
			),
			exp: match.NewSuffix("abc"),
		},
		{
			// #13
			tree: New(Pattern, nil,
				New(Text, TextData{"abc"}),
				New(Any, nil),
			),
			exp: match.NewPrefix("abc"),
		},
		{
			// #14
			tree: New(Pattern, nil,
				New(Text, TextData{"abc"}),
				New(Any, nil),
				New(Text, TextData{"def"}),
			),
			exp: match.NewPrefixSuffix("abc", "def"),
		},
		{
			// #15
			tree: New(Pattern, nil,
				New(Any, nil),
				New(Any, nil),
				New(Any, nil),
				New(Text, TextData{"abc"}),
				New(Any, nil),
				New(Any, nil),
			),
			exp: match.NewContains("abc"),
		},
		{
			// #16
			tree: New(Pattern, nil,
				New(Any, nil),
				New(Any, nil),
				New(Any, nil),
				New(Text, TextData{"abc"}),
				New(Any, nil),
				New(Any, nil),
			),
			sep: sep,
			exp: match.NewTree(
				match.NewText("abc"),
				match.NewAny(sep),
				match.NewAny(sep),
			),
		},
		{
			// #17
			// pattern: "**?abc**?"
			tree: New(Pattern, nil,
				New(Super, nil),
				New(Single, nil),
				New(Text, TextData{"abc"}),
				New(Super, nil),
				New(Single, nil),
			),
			exp: match.NewTree(
				match.NewText("abc"),
				match.NewMin(1),
				match.NewMin(1),
			),
		},
		{
			// #18
			tree: New(Pattern, nil,
				New(Text, TextData{"abc"}),
			),
			exp: match.NewText("abc"),
		},
		{
			// #19
			tree: New(Pattern, nil,
				New(AnyOf, nil,
					New(Pattern, nil,
						New(AnyOf, nil,
							New(Pattern, nil,
								New(Text, TextData{"abc"}),
							),
						),
					),
				),
			),
			exp: match.NewText("abc"),
		},
		{
			// #20
			tree: New(Pattern, nil,
				New(AnyOf, nil,
					New(Pattern, nil,
						New(Text, TextData{"abc"}),
						New(Single, nil),
					),
					New(Pattern, nil,
						New(Text, TextData{"abc"}),
						New(List, ListData{Chars: "def"}),
					),
					New(Pattern, nil,
						New(Text, TextData{"abc"}),
					),
					New(Pattern, nil,
						New(Text, TextData{"abc"}),
					),
				),
			),
			exp: match.NewTree(
				match.NewText("abc"),
				match.NothingMatcher{},
				match.NewAnyOf(
					match.NewSingle(nil),
					match.NewList([]rune{'d', 'e', 'f'}, false),
					match.NewNothing(),
				),
			),
		},
		{
			// #21
			tree: New(Pattern, nil,
				New(Range, RangeData{Lo: 'a', Hi: 'z'}),
				New(Range, RangeData{Lo: 'a', Hi: 'x', Not: true}),
				New(Any, nil),
			),
			exp: match.NewTree(
				match.NewRow([]match.MatchIndexSizer{
					match.NewRange('a', 'z', false),
					match.NewRange('a', 'x', true),
				}),
				match.NothingMatcher{},
				match.NewSuper(),
			),
		},
		{
			// #22
			tree: New(Pattern, nil,
				New(AnyOf, nil,
					New(Pattern, nil,
						New(Text, TextData{"abc"}),
						New(List, ListData{Chars: "abc"}),
						New(Text, TextData{"ghi"}),
					),
					New(Pattern, nil,
						New(Text, TextData{"abc"}),
						New(List, ListData{Chars: "def"}),
						New(Text, TextData{"ghi"}),
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
			m, err := test.tree.Match(test.sep)
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
