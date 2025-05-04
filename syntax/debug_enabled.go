//go:build globdebug

package syntax

import (
	"fmt"
	"os"
	"strings"
)

const debugEnabled = true

var (
	i      = 0
	prefix = map[int]string{}
)

func debugLogf(f string, args ...interface{}) {
	if f != "" && prefix[i] != "" {
		f = ": " + f
	}
	fmt.Fprint(os.Stderr,
		strings.Repeat("  ", i),
		fmt.Sprintf("(%d) ", i),
		prefix[i],
		fmt.Sprintf(f, args...),
		"\n",
	)
}

func debugIndexing(name, s string) func(int, []int) {
	EnterPrefix("%s: index: %q", name, s)
	return func(index int, segments []int) {
		Logf("-> %d, %v", index, segments)
		LeavePrefix()
	}
}

func debugMatching(name, s string) func(bool) {
	EnterPrefix("%s: match %q", name, s)
	return func(ok bool) {
		Logf("-> %t", ok)
		LeavePrefix()
	}
}

func debugEnterPrefix(s string, args ...interface{}) {
	Enter()
	prefix[i] = fmt.Sprintf(s, args...)
	Logf("")
}

func debugLeavePrefix() {
	prefix[i] = ""
	Leave()
}

func debugEnter() {
	i++
}

func debugLeave() {
	i--
}
