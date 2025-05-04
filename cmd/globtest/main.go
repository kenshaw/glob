package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/kenshaw/glob"
)

func main() {
	pattern := flag.String("p", "", "pattern to draw")
	sep := flag.String("s", "", "comma separated list of separators")
	fixture := flag.String("f", "", "fixture")
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()
	if err := run(*pattern, *sep, *fixture, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(pattern, sep, fixture string, verbose bool) error {
	if pattern == "" {
		return errors.New("pattern must not be empty")
	}
	var separators []rune
	for c := range strings.SplitSeq(sep, ",") {
		if r, w := utf8.DecodeRuneInString(c); len(c) > w {
			return errors.New("only single charactered separators are allowed")
		} else {
			separators = append(separators, r)
		}
	}
	g, err := glob.Compile(pattern, separators...)
	if err != nil {
		return fmt.Errorf("could not compile pattern: %w", err)
	}
	if !verbose {
		fmt.Println(g.Match(fixture))
		return nil
	}
	fmt.Printf("result: %t\n", g.Match(fixture))
	cb := testing.Benchmark(func(b *testing.B) {
		for b.Loop() {
			glob.Compile(pattern, separators...)
		}
	})
	fmt.Println("compile:", benchString(cb))
	mb := testing.Benchmark(func(b *testing.B) {
		for b.Loop() {
			g.Match(fixture)
		}
	})
	fmt.Println("match:    ", benchString(mb))
	return nil
}

func benchString(r testing.BenchmarkResult) string {
	nsop := r.NsPerOp()
	ns := fmt.Sprintf("%10d ns/op", nsop)
	allocs := "0"
	if r.N > 0 {
		if nsop < 100 {
			// The format specifiers here make sure that
			// the ones digits line up for all three possible formats.
			if nsop < 10 {
				ns = fmt.Sprintf("%13.2f ns/op", float64(r.T.Nanoseconds())/float64(r.N))
			} else {
				ns = fmt.Sprintf("%12.1f ns/op", float64(r.T.Nanoseconds())/float64(r.N))
			}
		}
		allocs = fmt.Sprintf("%d", r.MemAllocs/uint64(r.N))
	}
	return fmt.Sprintf("%8d\t%s\t%s allocs", r.N, ns, allocs)
}
