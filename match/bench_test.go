package match

import (
	"sync"
	"testing"
	"unicode/utf8"
)

func BenchmarkAppendMerge(b *testing.B) {
	s1 := []int{0, 1, 3, 6, 7}
	s2 := []int{0, 1, 3}
	for b.Loop() {
		appendMerge(s1, s2)
	}
}

func BenchmarkAppendMergeParallel(b *testing.B) {
	s1 := []int{0, 1, 3, 6, 7}
	s2 := []int{0, 1, 3}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			appendMerge(s1, s2)
		}
	})
}

func BenchmarkRuneLenFromTable(b *testing.B) {
	for b.Loop() {
		_ = table[runeToLen]
	}
}

func BenchmarkRuneLenFromUTF8(b *testing.B) {
	for b.Loop() {
		_ = utf8.RuneLen(runeToLen)
	}
}

func BenchmarkIndexContains(b *testing.B) {
	m := ContainsMatcher{string(bench_separators), true}
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexContainsParallel(b *testing.B) {
	m := ContainsMatcher{string(bench_separators), true}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkIndexList(b *testing.B) {
	m := NewList([]rune("def"), false)
	for b.Loop() {
		m.Index(bench_pattern)
	}
}

func BenchmarkIndexListParallel(b *testing.B) {
	m := NewList([]rune("def"), false)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Index(bench_pattern)
		}
	})
}

func BenchmarkIndexAny(b *testing.B) {
	m := NewAny(bench_separators)
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexAnyParallel(b *testing.B) {
	m := NewAny(bench_separators)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkIndexMax(b *testing.B) {
	m := NewMax(10)
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexMaxParallel(b *testing.B) {
	m := NewMax(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkIndexMin(b *testing.B) {
	m := NewMin(10)
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexMinParallel(b *testing.B) {
	m := NewMin(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkIndexNothing(b *testing.B) {
	m := NewNothing()
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexNothingParallel(b *testing.B) {
	m := NewNothing()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkIndexPrefixSuffix(b *testing.B) {
	m := NewPrefixSuffix("qew", "sqw")
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexPrefixSuffixParallel(b *testing.B) {
	m := NewPrefixSuffix("qew", "sqw")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkIndexPrefix(b *testing.B) {
	m := NewPrefix("qew")
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexPrefixParallel(b *testing.B) {
	m := NewPrefix("qew")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkIndexRange(b *testing.B) {
	m := NewRange('0', '9', false)
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexRangeParallel(b *testing.B) {
	m := NewRange('0', '9', false)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkRowIndex(b *testing.B) {
	m := NewRow([]MatchIndexSizer{
		NewText("abc"),
		NewText("def"),
		NewSingle(nil),
	})
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexRowParallel(b *testing.B) {
	m := NewRow([]MatchIndexSizer{
		NewText("abc"),
		NewText("def"),
		NewSingle(nil),
	})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkIndexSingle(b *testing.B) {
	m := NewSingle(bench_separators)
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexSingleParallel(b *testing.B) {
	m := NewSingle(bench_separators)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkIndexSuffix(b *testing.B) {
	m := NewSuffix("qwe")
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexSuffixParallel(b *testing.B) {
	m := NewSuffix("qwe")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkIndexSuper(b *testing.B) {
	m := NewSuper()
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexSuperParallel(b *testing.B) {
	m := NewSuper()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkIndexText(b *testing.B) {
	m := NewText("foo")
	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexTextParallel(b *testing.B) {
	m := NewText("foo")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}

func BenchmarkMatchTree(b *testing.B) {
	l := &fakeMatcher{4, 3, "left_fake"}
	r := &fakeMatcher{4, 3, "right_fake"}
	v := &fakeMatcher{2, 3, "value_fake"}
	// must be <= len(l + r + v)
	fixture := "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"
	bt := NewTree(v, l, r)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bt.Match(fixture)
		}
	})
}

func BenchmarkSegmentsPool_1(b *testing.B) {
	benchPool(1, b)
}

func BenchmarkSegmentsPool_2(b *testing.B) {
	benchPool(2, b)
}

func BenchmarkSegmentsPool_4(b *testing.B) {
	benchPool(4, b)
}

func BenchmarkSegmentsPool_8(b *testing.B) {
	benchPool(8, b)
}

func BenchmarkSegmentsPool_16(b *testing.B) {
	benchPool(16, b)
}

func BenchmarkSegmentsPool_32(b *testing.B) {
	benchPool(32, b)
}

func BenchmarkSegmentsPool_64(b *testing.B) {
	benchPool(64, b)
}

func BenchmarkSegmentsPool_128(b *testing.B) {
	benchPool(128, b)
}

func BenchmarkSegmentsPool_256(b *testing.B) {
	benchPool(256, b)
}

func BenchmarkSegmentsMake_1(b *testing.B) {
	benchMake(1, b)
}

func BenchmarkSegmentsMake_2(b *testing.B) {
	benchMake(2, b)
}

func BenchmarkSegmentsMake_4(b *testing.B) {
	benchMake(4, b)
}

func BenchmarkSegmentsMake_8(b *testing.B) {
	benchMake(8, b)
}

func BenchmarkSegmentsMake_16(b *testing.B) {
	benchMake(16, b)
}

func BenchmarkSegmentsMake_32(b *testing.B) {
	benchMake(32, b)
}

func BenchmarkSegmentsMake_64(b *testing.B) {
	benchMake(64, b)
}

func BenchmarkSegmentsMake_128(b *testing.B) {
	benchMake(128, b)
}

func BenchmarkSegmentsMake_256(b *testing.B) {
	benchMake(256, b)
}

func benchPool(i int, b *testing.B) {
	pool := sync.Pool{New: func() any {
		return make([]int, 0, i)
	}}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := pool.Get().([]int)[:0]
			pool.Put(s)
		}
	})
}

func benchMake(i int, b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = make([]int, 0, i)
		}
	})
}

type fakeMatcher struct {
	len  int
	segn int
	name string
}

func (f *fakeMatcher) Match(string) bool {
	return true
}

func (f *fakeMatcher) Index(s string) (int, []int) {
	seg := make([]int, 0, f.segn)
	for range f.segn {
		seg = append(seg, f.segn)
	}
	return 0, seg
}

func (f *fakeMatcher) Len() int {
	return f.len
}

// String satisfies the [fmt.Stringer] interface.
func (f *fakeMatcher) String() string {
	return f.name
}
