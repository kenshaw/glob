package glob

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// vector is one line of testdata/vectors.tsv (D19).
type vector struct {
	line                     int
	feature, pattern, s, sep string
	expected, source         string
}

// unsupported names the features of the survey that the fork does not hold,
// with the reason. TestVectors skips their vectors. docs/SYNTAX.md holds the
// survey.
var unsupported = map[string]string{
	"opt-period":             "the fork has no option for a leading dot",
	"opt-leadingdir":         "the fork has no option FNM_LEADING_DIR",
	"opt-noescape":           "the fork has no option FNM_NOESCAPE",
	"opt-negate":             "the fork has no negation of a whole pattern",
	"opt-matchbase":          "the fork has no option to match a base name",
	"dialect-editorconfig":   "editorconfig is another dialect",
	"dialect-gitignore":      "gitignore is another dialect",
	"dialect-python-fnmatch": "Python fnmatch is another dialect",
}

// readVectors reads testdata/vectors.tsv. A field that starts with a double
// quote is a quoted Go string.
func readVectors(t *testing.T) []vector {
	t.Helper()
	f, err := os.Open("testdata/vectors.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []vector
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 1<<20)
	for line := 1; sc.Scan(); line++ {
		if line == 1 {
			if got, exp := sc.Text(), "feature\tpattern\tinput\tseparators\texpected\tsource"; got != exp {
				t.Fatalf("testdata/vectors.tsv: expected the header %q, got: %q", exp, got)
			}
			continue
		}
		cols := strings.Split(sc.Text(), "\t")
		if len(cols) != 6 {
			t.Fatalf("testdata/vectors.tsv:%d: expected 6 columns, got: %d", line, len(cols))
		}
		for i, c := range cols[:4] {
			if strings.HasPrefix(c, `"`) {
				s, err := strconv.Unquote(c)
				if err != nil {
					t.Fatalf("testdata/vectors.tsv:%d: unquoting %s: %v", line, c, err)
				}
				cols[i] = s
			}
		}
		out = append(out, vector{line, cols[0], cols[1], cols[2], cols[3], cols[4], cols[5]})
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// TestVectors holds D19. It runs each vector of the survey with the optimized
// matchers and with the backtracking matcher (D12).
func TestVectors(t *testing.T) {
	t.Parallel()
	var ran int
	for _, v := range readVectors(t) {
		if _, ok := unsupported[v.feature]; ok {
			continue
		}
		sep := []rune(v.sep)
		opts := []Option{WithSeparators(sep...)}
		switch v.feature {
		case "opt-globstar":
			opts = append(opts, WithGlobstar())
		case "opt-nocase":
			opts = append(opts, WithFold())
		}
		ran++
		g, err := CompileWith(v.pattern, opts...)
		switch {
		case v.expected == "error" && err == nil:
			t.Errorf("vectors.tsv:%d: %s: CompileWith(%q): expected an error, got: nil (%s)", v.line, v.feature, v.pattern, v.source)
			continue
		case v.expected == "error":
			continue
		case err != nil:
			t.Errorf("vectors.tsv:%d: %s: CompileWith(%q): expected no error, got: %v (%s)", v.line, v.feature, v.pattern, err, v.source)
			continue
		}
		exp := v.expected == "true"
		if got := g.Match(v.s); got != exp {
			t.Errorf("vectors.tsv:%d: %s: %q, %q, separators %q: expected %t, got: %t (%s)", v.line, v.feature, v.pattern, v.s, v.sep, exp, got, v.source)
		}
		if v.feature == "opt-globstar" || v.feature == "opt-nocase" {
			continue
		}
		ref, ok := backtrack(t, v.pattern, sep)
		if !ok {
			t.Errorf("vectors.tsv:%d: Parse(%q) failed and Compile did not", v.line, v.pattern)
			continue
		}
		if got := ref.Match(v.s); got != exp {
			t.Errorf("vectors.tsv:%d: %s: backtracking: %q, %q, separators %q: expected %t, got: %t (%s)", v.line, v.feature, v.pattern, v.s, v.sep, exp, got, v.source)
		}
	}
	if ran < 2000 {
		t.Errorf("expected at least 2000 vectors to run, got: %d", ran)
	}
}
