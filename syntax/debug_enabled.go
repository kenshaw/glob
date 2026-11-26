//go:build globdebug

package syntax

import (
	"fmt"
	"os"
	"strings"
)

const debugEnabled = true

var (
	depth  = 0
	prefix = map[int]string{}
)

func debugLogf(f string, args ...any) {
	if f != "" && prefix[depth] != "" {
		f = ": " + f
	}
	fmt.Fprint(os.Stderr,
		strings.Repeat("  ", depth),
		fmt.Sprintf("(%d) ", depth),
		prefix[depth],
		fmt.Sprintf(f, args...),
		"\n",
	)
}

func debugIndexing(name, s string) func(int, []int) {
	debugEnterPrefix("%s: index: %q", name, s)
	return func(index int, segments []int) {
		debugLogf("-> %d, %v", index, segments)
		debugLeavePrefix()
	}
}

func debugMatching(name, s string) func(bool) {
	debugEnterPrefix("%s: match %q", name, s)
	return func(ok bool) {
		debugLogf("-> %t", ok)
		debugLeavePrefix()
	}
}

func debugEnterPrefix(s string, args ...any) {
	debugEnter()
	prefix[depth] = fmt.Sprintf(s, args...)
	debugLogf("")
}

func debugLeavePrefix() {
	prefix[depth] = ""
	debugLeave()
}

func debugEnter() {
	depth++
}

func debugLeave() {
	depth--
}
