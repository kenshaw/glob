//go:build !race

package glob

// race reports whether the race detector is on. See race_test.go.
const race = false
