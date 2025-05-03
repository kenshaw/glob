package glob_test

import (
	"github.com/kenshaw/glob"
)

func main() {
	// create simple glob
	g := glob.MustCompile("*.github.com")
	g.Match("api.github.com") // true

	// quote meta characters and then create simple glob
	g = glob.MustCompile(glob.QuoteMeta("*.github.com"))
	g.Match("*.github.com") // true

	// create new glob with set of delimiters as ["."]
	g = glob.MustCompile("api.*.com", '.')
	g.Match("api.github.com") // true
	g.Match("api.gi.hub.com") // false

	// create new glob with set of delimiters as ["."]
	// but now with super wildcard
	g = glob.MustCompile("api.**.com", '.')
	g.Match("api.github.com") // true
	g.Match("api.gi.hub.com") // true

	// create glob with single symbol wildcard
	g = glob.MustCompile("?at")
	g.Match("cat") // true
	g.Match("fat") // true
	g.Match("at")  // false

	// create glob with single symbol wildcard and delimiters ['f']
	g = glob.MustCompile("?at", 'f')
	g.Match("cat") // true
	g.Match("fat") // false
	g.Match("at")  // false

	// create glob with character-list matchers
	g = glob.MustCompile("[abc]at")
	g.Match("cat") // true
	g.Match("bat") // true
	g.Match("fat") // false
	g.Match("at")  // false

	// create glob with character-list matchers
	g = glob.MustCompile("[!abc]at")
	g.Match("cat") // false
	g.Match("bat") // false
	g.Match("fat") // true
	g.Match("at")  // false

	// create glob with character-range matchers
	g = glob.MustCompile("[a-c]at")
	g.Match("cat") // true
	g.Match("bat") // true
	g.Match("fat") // false
	g.Match("at")  // false

	// create glob with character-range matchers
	g = glob.MustCompile("[!a-c]at")
	g.Match("cat") // false
	g.Match("bat") // false
	g.Match("fat") // true
	g.Match("at")  // false

	// create glob with pattern-alternatives list
	g = glob.MustCompile("{cat,bat,[fr]at}")
	g.Match("cat")  // true
	g.Match("bat")  // true
	g.Match("fat")  // true
	g.Match("rat")  // true
	g.Match("at")   // false
	g.Match("zat")  // false
	g.Match("frat") // false
	// Output:
}
