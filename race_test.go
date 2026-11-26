//go:build race

package glob

// race reports whether the race detector is on. With it, [sync.Pool] drops
// some of the values that it gets, so a match can allocate.
const race = true
