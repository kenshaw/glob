package syntax

import (
	"reflect"
	"strconv"
	"testing"
)

func TestNode(t *testing.T) {
	sep := []rune{'.'}
	for i, test := range []struct {
		tree *Node
		sep  []rune
		exp  Matcher
	}{
		{
			// #0
			tree: New(Pattern, nil,
				New(Text, TextData{"abc"}),
			),
			exp: NewText("abc"),
		},
		{
			// #1
			tree: New(Pattern, nil,
				New(Any, nil),
			),
			sep: sep,
			exp: NewAny(sep),
		},
		{
			// #2
			tree: New(Pattern, nil,
				New(Any, nil),
			),
			exp: NewSuper(),
		},
		{
			// #3
			tree: New(Pattern, nil,
				New(Super, nil),
			),
			exp: NewSuper(),
		},
		{
			// #4
			tree: New(Pattern, nil,
				New(Single, nil),
			),
			sep: sep,
			exp: NewSingle(sep),
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
			exp: NewRange('a', 'z', true),
		},
		{
			// #6
			tree: New(Pattern, nil,
				New(List, ListData{
					Chars: "abc",
					Not:   true,
				}),
			),
			exp: NewList([]rune{'a', 'b', 'c'}, true),
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
			exp: NewEveryOf([]Matcher{
				NewMin(3),
				NewAny(sep),
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
			exp: NewMin(3),
		},
		{
			// #9
			tree: New(Pattern, nil,
				New(Any, nil),
				New(Text, TextData{"abc"}),
				New(Single, nil),
			),
			sep: sep,
			exp: NewTree(
				NewRow([]MatchIndexSizer{
					NewText("abc"),
					NewSingle(sep),
				}),
				NewAny(sep),
				NothingMatcher{},
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
			exp: NewTree(
				NewText("/"),
				NothingMatcher{},
				NewTree(
					MustIndexedAnyOf(
						NewText("z"),
						NewText("ab"),
					),
					NothingMatcher{},
					NewSuper(),
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
			exp: NewTree(
				NewRow([]MatchIndexSizer{
					NewSingle(sep),
					NewText("abc"),
					NewSingle(sep),
				}),
				NewSuper(),
				NothingMatcher{},
			),
		},
		{
			// #12
			tree: New(Pattern, nil,
				New(Any, nil),
				New(Text, TextData{"abc"}),
			),
			exp: NewSuffix("abc"),
		},
		{
			// #13
			tree: New(Pattern, nil,
				New(Text, TextData{"abc"}),
				New(Any, nil),
			),
			exp: NewPrefix("abc"),
		},
		{
			// #14
			tree: New(Pattern, nil,
				New(Text, TextData{"abc"}),
				New(Any, nil),
				New(Text, TextData{"def"}),
			),
			exp: NewPrefixSuffix("abc", "def"),
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
			exp: NewContains("abc"),
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
			exp: NewTree(
				NewText("abc"),
				NewAny(sep),
				NewAny(sep),
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
			exp: NewTree(
				NewText("abc"),
				NewMin(1),
				NewMin(1),
			),
		},
		{
			// #18
			tree: New(Pattern, nil,
				New(Text, TextData{"abc"}),
			),
			exp: NewText("abc"),
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
			exp: NewText("abc"),
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
			exp: NewTree(
				NewText("abc"),
				NothingMatcher{},
				NewAnyOf(
					NewSingle(nil),
					NewList([]rune{'d', 'e', 'f'}, false),
					NewNothing(),
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
			exp: NewTree(
				NewRow([]MatchIndexSizer{
					NewRange('a', 'z', false),
					NewRange('a', 'x', true),
				}),
				NothingMatcher{},
				NewSuper(),
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
			exp: NewRow([]MatchIndexSizer{
				NewText("abc"),
				MustIndexedSizedAnyOf(
					NewList([]rune{'a', 'b', 'c'}, false),
					NewList([]rune{'d', 'e', 'f'}, false),
				),
				NewText("ghi"),
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
					Graphviz("act", m.(Matcher)),
					Graphviz("exp", test.exp.(Matcher)),
				)
			}
		})
	}
}
