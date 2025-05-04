package match

import (
	"reflect"
	"strconv"
	"testing"
)

var separators = []rune{'.'}

func TestCompile(t *testing.T) {
	for i, test := range []struct {
		m   []Matcher
		exp Matcher
	}{
		{
			[]Matcher{
				NewSuper(),
				NewSingle(nil),
			},
			NewMin(1),
		},
		{
			[]Matcher{
				NewAny(separators),
				NewSingle(separators),
			},
			NewEveryOf([]Matcher{
				NewMin(1),
				NewAny(separators),
			}),
		},
		{
			[]Matcher{
				NewSingle(nil),
				NewSingle(nil),
				NewSingle(nil),
			},
			NewEveryOf([]Matcher{
				NewMin(3),
				NewMax(3),
			}),
		},
		{
			[]Matcher{
				NewList([]rune{'a'}, true),
				NewAny([]rune{'a'}),
			},
			NewEveryOf([]Matcher{
				NewMin(1),
				NewAny([]rune{'a'}),
			}),
		},
		{
			[]Matcher{
				NewSuper(),
				NewSingle(separators),
				NewText("c"),
			},
			NewTree(
				NewText("c"),
				NewTree(
					NewSingle(separators),
					NewSuper(),
					NothingMatcher{},
				),
				NothingMatcher{},
			),
		},
		{
			[]Matcher{
				NewAny(nil),
				NewText("c"),
				NewAny(nil),
			},
			NewTree(
				NewText("c"),
				NewAny(nil),
				NewAny(nil),
			),
		},
		{
			[]Matcher{
				NewRange('a', 'c', true),
				NewList([]rune{'z', 't', 'e'}, false),
				NewText("c"),
				NewSingle(nil),
			},
			NewRow([]MatchIndexSizer{
				NewRange('a', 'c', true),
				NewList([]rune{'z', 't', 'e'}, false),
				NewText("c"),
				NewSingle(nil),
			}),
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			m, err := Compile(test.m)
			if err != nil {
				t.Fatalf("Compile() error: %s", err)
			}
			if !reflect.DeepEqual(m, test.exp) {
				t.Errorf(
					"Compile():\nact: %#v;\nexp: %#v;\ngraphviz:\n%s\n%s",
					m, test.exp,
					Graphviz("act", m), Graphviz("exp", test.exp),
				)
			}
		})
	}
}

func TestMinimize(t *testing.T) {
	for i, test := range []struct {
		m, exp []Matcher
	}{
		{
			m: []Matcher{
				NewRange('a', 'c', true),
				NewList([]rune{'z', 't', 'e'}, false),
				NewText("c"),
				NewSingle(nil),
				NewAny(nil),
			},
			exp: []Matcher{
				NewRow([]MatchIndexSizer{
					NewRange('a', 'c', true),
					NewList([]rune{'z', 't', 'e'}, false),
					NewText("c"),
				}),
				NewMin(1),
			},
		},
		{
			m: []Matcher{
				NewRange('a', 'c', true),
				NewList([]rune{'z', 't', 'e'}, false),
				NewText("c"),
				NewSingle(nil),
				NewAny(nil),
				NewSingle(nil),
				NewSingle(nil),
				NewAny(nil),
			},
			exp: []Matcher{
				NewRow([]MatchIndexSizer{
					NewRange('a', 'c', true),
					NewList([]rune{'z', 't', 'e'}, false),
					NewText("c"),
				}),
				NewMin(3),
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			act := Minimize(test.m)
			if !reflect.DeepEqual(act, test.exp) {
				t.Errorf(
					"Minimize():\nact: %#v;\nexp: %#v",
					act, test.exp,
				)
			}
		})
	}
}
