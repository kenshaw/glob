package glob_test

import (
	"strings"
	"testing"
)

// TestTheModuleHasNoRequire holds D5. go.mod requires no module, so a program
// that imports glob gets the standard library and nothing else.
func TestTheModuleHasNoRequire(t *testing.T) {
	t.Parallel()
	for i, line := range strings.Split(read(t, "go.mod"), "\n") {
		if fields := strings.Fields(line); len(fields) != 0 && fields[0] == "require" {
			t.Errorf("go.mod:%d: %q. The module requires no module. See D5", i+1, line)
		}
	}
}
