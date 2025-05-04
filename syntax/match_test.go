package syntax

import (
	"reflect"
	"strconv"
	"testing"
	"unicode/utf8"
)

var bench_separators = []rune{'.'}

const bench_pattern = "abcdefghijklmnopqrstuvwxyz0123456789"

func TestAppendMerge(t *testing.T) {
	for id, test := range []struct {
		segments [2][]int
		exp      []int
	}{
		{
			[2][]int{
				{0, 6, 7},
				{0, 1, 3},
			},
			[]int{0, 1, 3, 6, 7},
		},
		{
			[2][]int{
				{0, 1, 3, 6, 7},
				{0, 1, 10},
			},
			[]int{0, 1, 3, 6, 7, 10},
		},
	} {
		act := appendMerge(test.segments[0], test.segments[1])
		if !reflect.DeepEqual(act, test.exp) {
			t.Errorf("#%d merge sort segments unexpected:\nact: %v\nexp:%v", id, act, test.exp)
			continue
		}
	}
}

func TestCompile(t *testing.T) {
	sep := []rune{'.'}
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
				NewAny(sep),
				NewSingle(sep),
			},
			NewEveryOf([]Matcher{
				NewMin(1),
				NewAny(sep),
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
				NewSingle(sep),
				NewText("c"),
			},
			NewTree(
				NewText("c"),
				NewTree(
					NewSingle(sep),
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
			m, err := BuildMatcher(test.m)
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
			act := MinimizeMatcher(test.m)
			if !reflect.DeepEqual(act, test.exp) {
				t.Errorf(
					"Minimize():\nact: %#v;\nexp: %#v",
					act, test.exp,
				)
			}
		})
	}
}

func getTable() []int {
	table := make([]int, utf8.MaxRune+1)
	for i := 0; i <= utf8.MaxRune; i++ {
		table[i] = utf8.RuneLen(rune(i))
	}
	return table
}

var table = getTable()

const runeToLen = 'q'
