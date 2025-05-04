//go:build !globdebug

package syntax

const debugEnabled = false

func debugLogf(string, ...any)        {}
func debugEnter()                     {}
func debugLeave()                     {}
func debugEnterPrefix(string, ...any) {}
func debugLeavePrefix()               {}
func debugIndexing(n, s string) func(int, []int) {
	panic("must never be called")
}

func debugMatching(n, s string) func(bool) {
	panic("must never be called")
}
