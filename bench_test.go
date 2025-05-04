package glob

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/kenshaw/glob/match"
)

func BenchmarkParseGlob(b *testing.B) {
	for b.Loop() {
		Compile(pattern_all)
	}
}

func BenchmarkParseRegexp(b *testing.B) {
	for b.Loop() {
		regexp.MustCompile(regexp_all)
	}
}

func BenchmarkAllGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_all)
	for b.Loop() {
		_ = m.Match(fixture_all_match)
	}
}

func BenchmarkAllRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_all)
	f := []byte(fixture_all_match)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkAllGlobMismatch(b *testing.B) {
	g := MustCompile(pattern_all)
	fmt.Println(match.Graphviz(pattern_all, g))
	for b.Loop() {
		_ = g.Match(fixture_all_mismatch)
	}
}

func BenchmarkAllRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_all)
	f := []byte(fixture_all_mismatch)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkMultipleGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_multiple)
	for b.Loop() {
		_ = m.Match(fixture_multiple_match)
	}
}

func BenchmarkMultipleRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_multiple)
	f := []byte(fixture_multiple_match)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkMultipleGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_multiple)
	for b.Loop() {
		_ = m.Match(fixture_multiple_mismatch)
	}
}

func BenchmarkMultipleRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_multiple)
	f := []byte(fixture_multiple_mismatch)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives)
	for b.Loop() {
		_ = m.Match(fixture_alternatives_match)
	}
}

func BenchmarkAlternativesGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives)
	for b.Loop() {
		_ = m.Match(fixture_alternatives_mismatch)
	}
}

func BenchmarkAlternativesRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives)
	f := []byte(fixture_alternatives_match)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives)
	f := []byte(fixture_alternatives_mismatch)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesSuffixFirstGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_suffix)
	for b.Loop() {
		_ = m.Match(fixture_alternatives_suffix_first_match)
	}
}

func BenchmarkAlternativesSuffixFirstGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_suffix)
	for b.Loop() {
		_ = m.Match(fixture_alternatives_suffix_first_mismatch)
	}
}

func BenchmarkAlternativesSuffixSecondGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_suffix)
	for b.Loop() {
		_ = m.Match(fixture_alternatives_suffix_second)
	}
}

func BenchmarkAlternativesCombineLiteGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_combine_lite)
	for b.Loop() {
		_ = m.Match(fixture_alternatives_combine_lite)
	}
}

func BenchmarkAlternativesCombineHardGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_combine_hard)
	for b.Loop() {
		_ = m.Match(fixture_alternatives_combine_hard)
	}
}

func BenchmarkAlternativesSuffixFirstRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_suffix)
	f := []byte(fixture_alternatives_suffix_first_match)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesSuffixFirstRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_suffix)
	f := []byte(fixture_alternatives_suffix_first_mismatch)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesSuffixSecondRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_suffix)
	f := []byte(fixture_alternatives_suffix_second)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesCombineLiteRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_combine_lite)
	f := []byte(fixture_alternatives_combine_lite)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesCombineHardRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_combine_hard)
	f := []byte(fixture_alternatives_combine_hard)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkPlainGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_plain)
	for b.Loop() {
		_ = m.Match(fixture_plain_match)
	}
}

func BenchmarkPlainRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_plain)
	f := []byte(fixture_plain_match)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkPlainGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_plain)
	for b.Loop() {
		_ = m.Match(fixture_plain_mismatch)
	}
}

func BenchmarkPlainRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_plain)
	f := []byte(fixture_plain_mismatch)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkPrefixGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_prefix)
	for b.Loop() {
		_ = m.Match(fixture_prefix_suffix_match)
	}
}

func BenchmarkPrefixRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix)
	f := []byte(fixture_prefix_suffix_match)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkPrefixGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_prefix)
	for b.Loop() {
		_ = m.Match(fixture_prefix_suffix_mismatch)
	}
}

func BenchmarkPrefixRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix)
	f := []byte(fixture_prefix_suffix_mismatch)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkSuffixGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_suffix)
	for b.Loop() {
		_ = m.Match(fixture_prefix_suffix_match)
	}
}

func BenchmarkSuffixRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_suffix)
	f := []byte(fixture_prefix_suffix_match)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkSuffixGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_suffix)
	for b.Loop() {
		_ = m.Match(fixture_prefix_suffix_mismatch)
	}
}

func BenchmarkSuffixRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_suffix)
	f := []byte(fixture_prefix_suffix_mismatch)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkPrefixSuffixGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_prefix_suffix)
	for b.Loop() {
		_ = m.Match(fixture_prefix_suffix_match)
	}
}

func BenchmarkPrefixSuffixRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix_suffix)
	f := []byte(fixture_prefix_suffix_match)
	for b.Loop() {
		_ = m.Match(f)
	}
}

func BenchmarkPrefixSuffixGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_prefix_suffix)
	for b.Loop() {
		_ = m.Match(fixture_prefix_suffix_mismatch)
	}
}

func BenchmarkPrefixSuffixRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix_suffix)
	f := []byte(fixture_prefix_suffix_mismatch)
	for b.Loop() {
		_ = m.Match(f)
	}
}
