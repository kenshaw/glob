package syntax

import (
	"fmt"
	"reflect"
	"testing"
)

func TestIndexedAnyOf(t *testing.T) {
	for id, test := range []struct {
		matchers Matchers
		fixture  string
		index    int
		segments []int
	}{
		{
			Matchers{
				NewAny(nil),
				NewText("b"),
				NewText("c"),
			},
			"abc",
			0,
			[]int{0, 1, 2, 3},
		},
		{
			Matchers{
				NewPrefix("b"),
				NewSuffix("c"),
			},
			"abc",
			0,
			[]int{3},
		},
		{
			Matchers{
				NewList([]rune("[def]"), false),
				NewList([]rune("[abc]"), false),
			},
			"abcdef",
			0,
			[]int{1},
		},
	} {
		t.Run("", func(t *testing.T) {
			a := NewAnyOf(test.matchers...).(Indexer)
			index, segments := a.Index(test.fixture)
			if index != test.index {
				t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
			}
			if !reflect.DeepEqual(segments, test.segments) {
				t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
			}
		})
	}
}

func TestAnyIndex(t *testing.T) {
	for id, test := range []struct {
		sep      []rune
		fixture  string
		index    int
		segments []int
	}{
		{
			[]rune{'.'},
			"abc",
			0,
			[]int{0, 1, 2, 3},
		},
		{
			[]rune{'.'},
			"abc.def",
			0,
			[]int{0, 1, 2, 3},
		},
	} {
		p := NewAny(test.sep)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestContainsIndex(t *testing.T) {
	for id, test := range []struct {
		prefix   string
		not      bool
		fixture  string
		index    int
		segments []int
	}{
		{
			"ab",
			false,
			"abc",
			0,
			[]int{2, 3},
		},
		{
			"ab",
			false,
			"fffabfff",
			0,
			[]int{5, 6, 7, 8},
		},
		{
			"ab",
			true,
			"abc",
			0,
			[]int{0},
		},
		{
			"ab",
			true,
			"fffabfff",
			0,
			[]int{0, 1, 2, 3},
		},
	} {
		p := ContainsMatcher{test.prefix, test.not}
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestIndexedEveryOf(t *testing.T) {
	for id, test := range []struct {
		matchers Matchers
		fixture  string
		index    int
		segments []int
	}{
		{
			Matchers{
				NewAny(nil),
				NewText("b"),
				NewText("c"),
			},
			"dbc",
			-1,
			nil,
		},
		{
			Matchers{
				NewAny(nil),
				NewPrefix("b"),
				NewSuffix("c"),
			},
			"abc",
			1,
			[]int{2},
		},
	} {
		everyOf := NewEveryOf(test.matchers).(IndexedEveryOf)
		index, segments := everyOf.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestListIndex(t *testing.T) {
	for id, test := range []struct {
		list     []rune
		not      bool
		fixture  string
		index    int
		segments []int
	}{
		{
			[]rune("ab"),
			false,
			"abc",
			0,
			[]int{1},
		},
		{
			[]rune("ab"),
			true,
			"fffabfff",
			0,
			[]int{1},
		},
	} {
		p := NewList(test.list, test.not)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestMaxIndex(t *testing.T) {
	for id, test := range []struct {
		limit    int
		fixture  string
		index    int
		segments []int
	}{
		{
			3,
			"abc",
			0,
			[]int{0, 1, 2, 3},
		},
		{
			3,
			"abcdef",
			0,
			[]int{0, 1, 2, 3},
		},
	} {
		p := NewMax(test.limit)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestMinIndex(t *testing.T) {
	for id, test := range []struct {
		limit    int
		fixture  string
		index    int
		segments []int
	}{
		{
			1,
			"abc",
			0,
			[]int{1, 2, 3},
		},
		{
			3,
			"abcd",
			0,
			[]int{3, 4},
		},
	} {
		p := NewMin(test.limit)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestNothingIndex(t *testing.T) {
	for id, test := range []struct {
		fixture  string
		index    int
		segments []int
	}{
		{
			"abc",
			0,
			[]int{0},
		},
		{
			"",
			0,
			[]int{0},
		},
	} {
		p := NewNothing()
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestPrefixAnyIndex(t *testing.T) {
	for id, test := range []struct {
		prefix     string
		separators []rune
		fixture    string
		index      int
		segments   []int
	}{
		{
			"ab",
			[]rune{'.'},
			"ab",
			0,
			[]int{2},
		},
		{
			"ab",
			[]rune{'.'},
			"abc",
			0,
			[]int{2, 3},
		},
		{
			"ab",
			[]rune{'.'},
			"qw.abcd.efg",
			3,
			[]int{2, 3, 4},
		},
	} {
		p := NewPrefixAny(test.prefix, test.separators)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestPrefixSuffixIndex(t *testing.T) {
	for id, test := range []struct {
		prefix   string
		suffix   string
		fixture  string
		index    int
		segments []int
	}{
		{
			"a",
			"c",
			"abc",
			0,
			[]int{3},
		},
		{
			"f",
			"f",
			"fffabfff",
			0,
			[]int{1, 2, 3, 6, 7, 8},
		},
		{
			"ab",
			"bc",
			"abc",
			0,
			[]int{3},
		},
	} {
		p := NewPrefixSuffix(test.prefix, test.suffix)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestRangeIndex(t *testing.T) {
	for id, test := range []struct {
		lo, hi   rune
		not      bool
		fixture  string
		index    int
		segments []int
	}{
		{
			'a', 'z',
			false,
			"abc",
			0,
			[]int{1},
		},
		{
			'a', 'c',
			false,
			"abcd",
			0,
			[]int{1},
		},
		{
			'a', 'c',
			true,
			"abcd",
			3,
			[]int{1},
		},
	} {
		m := NewRange(test.lo, test.hi, test.not)
		index, segments := m.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestRowIndex(t *testing.T) {
	for id, test := range []struct {
		matchers []MatchIndexSizer
		fixture  string
		index    int
		segments []int
	}{
		{
			[]MatchIndexSizer{
				NewText("abc"),
				NewText("def"),
				NewSingle(nil),
			},
			"qweabcdefghij",
			3,
			[]int{7},
		},
		{
			[]MatchIndexSizer{
				NewText("abc"),
				NewText("def"),
				NewSingle(nil),
			},
			"abcd",
			-1,
			nil,
		},
	} {
		p := NewRow(test.matchers)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestSingleIndex(t *testing.T) {
	for id, test := range []struct {
		separators []rune
		fixture    string
		index      int
		segments   []int
	}{
		{
			[]rune{'.'},
			".abc",
			1,
			[]int{1},
		},
		{
			[]rune{'.'},
			".",
			-1,
			nil,
		},
	} {
		p := NewSingle(test.separators)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestSuffixAnyIndex(t *testing.T) {
	for id, test := range []struct {
		suffix     string
		separators []rune
		fixture    string
		index      int
		segments   []int
	}{
		{
			"ab",
			[]rune{'.'},
			"ab",
			0,
			[]int{2},
		},
		{
			"ab",
			[]rune{'.'},
			"cab",
			0,
			[]int{3},
		},
		{
			"ab",
			[]rune{'.'},
			"qw.cdab.efg",
			3,
			[]int{4},
		},
	} {
		p := NewSuffixAny(test.suffix, test.separators)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestSuffixIndex(t *testing.T) {
	for id, test := range []struct {
		prefix   string
		fixture  string
		index    int
		segments []int
	}{
		{
			"ab",
			"abc",
			0,
			[]int{2},
		},
		{
			"ab",
			"fffabfff",
			0,
			[]int{5},
		},
	} {
		p := NewSuffix(test.prefix)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestTextIndex(t *testing.T) {
	for id, test := range []struct {
		text     string
		fixture  string
		index    int
		segments []int
	}{
		{
			"b",
			"abc",
			1,
			[]int{1},
		},
		{
			"f",
			"abcd",
			-1,
			nil,
		},
	} {
		m := NewText(test.text)
		index, segments := m.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestTree(t *testing.T) {
	for _, test := range []struct {
		tree Matcher
		str  string
		exp  bool
	}{
		{
			NewTree(NewText("x"), NewText("y"), NewText("z")),
			"0000x1111",
			false,
		},
		{
			NewTree(NewText("a"), NewSingle(nil), NewSingle(nil)),
			"aaa",
			true,
		},
		{
			NewTree(NewText("abc"), NewSuper(), NewSuper()),
			"abc",
			true,
		},
		{
			NewTree(NewText("a"), NewSingle(nil), NewSingle(nil)),
			"aaa",
			true,
		},
		{
			NewTree(NewText("b"), NewSingle(nil), NothingMatcher{}),
			"bbb",
			false,
		},
		{
			NewTree(
				NewText("c"),
				NewTree(
					NewSingle(nil),
					NewSuper(),
					NothingMatcher{},
				),
				NothingMatcher{},
			),
			"abc",
			true,
		},
	} {
		t.Run("", func(t *testing.T) {
			act := test.tree.Match(test.str)
			if act != test.exp {
				fmt.Println(Graphviz("NIL", test.tree))
				t.Errorf("match %q error: act: %t; exp: %t", test.str, act, test.exp)
			}
		})
	}
}

func TestPrefixIndex(t *testing.T) {
	for id, test := range []struct {
		prefix   string
		fixture  string
		index    int
		segments []int
	}{
		{
			"ab",
			"abc",
			0,
			[]int{2, 3},
		},
		{
			"ab",
			"fffabfff",
			3,
			[]int{2, 3, 4, 5},
		},
	} {
		p := NewPrefix(test.prefix)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestSuperIndex(t *testing.T) {
	for id, test := range []struct {
		fixture  string
		index    int
		segments []int
	}{
		{
			"abc",
			0,
			[]int{0, 1, 2, 3},
		},
		{
			"",
			0,
			[]int{0},
		},
	} {
		p := NewSuper()
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}
