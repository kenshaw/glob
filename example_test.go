package glob_test

import (
	"fmt"

	"github.com/kenshaw/glob"
)

func Example() {
	// create simple glob
	g1 := glob.Must("*.github.com")
	fmt.Println(g1)
	fmt.Println(g1.Match("api.github.com")) // true

	// quote meta characters and then create simple glob
	g2 := glob.Must(glob.Quote("*.github.com"))
	fmt.Println(g2)
	fmt.Println(g2.Match("*.github.com")) // true

	// create new glob with set of delimiters as ["."]
	g3 := glob.Must("api.*.com", '.')
	fmt.Println(g3)
	fmt.Println(g3.Match("api.github.com")) // true
	fmt.Println(g3.Match("api.gi.hub.com")) // false

	// create new glob with set of delimiters as ["."]
	// but now with super wildcard
	g4 := glob.Must("api.**.com", '.')
	fmt.Println(g4)
	fmt.Println(g4.Match("api.github.com")) // true
	fmt.Println(g4.Match("api.gi.hub.com")) // true

	// create glob with single symbol wildcard
	g5 := glob.Must("?at")
	fmt.Println(g5)
	fmt.Println(g5.Match("cat")) // true
	fmt.Println(g5.Match("fat")) // true
	fmt.Println(g5.Match("at"))  // false

	// create glob with single symbol wildcard and delimiters ['f']
	g6 := glob.Must("?at", 'f')
	fmt.Println(g6)
	fmt.Println(g6.Match("cat")) // true
	fmt.Println(g6.Match("fat")) // false
	fmt.Println(g6.Match("at"))  // false

	// create glob with character-list matchers
	g7 := glob.Must("[abc]at")
	fmt.Println(g7)
	fmt.Println(g7.Match("cat")) // true
	fmt.Println(g7.Match("bat")) // true
	fmt.Println(g7.Match("fat")) // false
	fmt.Println(g7.Match("at"))  // false

	// create glob with character-list matchers
	g8 := glob.Must("[!abc]at")
	fmt.Println(g8)
	fmt.Println(g8.Match("cat")) // false
	fmt.Println(g8.Match("bat")) // false
	fmt.Println(g8.Match("fat")) // true
	fmt.Println(g8.Match("at"))  // false

	// create glob with character-range matchers
	g9 := glob.Must("[a-c]at")
	fmt.Println(g9)
	fmt.Println(g9.Match("cat")) // true
	fmt.Println(g9.Match("bat")) // true
	fmt.Println(g9.Match("fat")) // false
	fmt.Println(g9.Match("at"))  // false

	// create glob with character-range matchers
	g10 := glob.Must("[!a-c]at")
	fmt.Println(g10)
	fmt.Println(g10.Match("cat")) // false
	fmt.Println(g10.Match("bat")) // false
	fmt.Println(g10.Match("fat")) // true
	fmt.Println(g10.Match("at"))  // false

	// create glob with pattern-alternatives list
	g11 := glob.Must("{cat,bat,[fr]at}")
	fmt.Println(g11)
	fmt.Println(g11.Match("cat"))  // true
	fmt.Println(g11.Match("bat"))  // true
	fmt.Println(g11.Match("fat"))  // true
	fmt.Println(g11.Match("rat"))  // true
	fmt.Println(g11.Match("at"))   // false
	fmt.Println(g11.Match("zat"))  // false
	fmt.Println(g11.Match("frat")) // false

	// Output:
	// <suffix:.github.com>
	// true
	// <text:`*.github.com`>
	// true
	// <btree:[<nothing><-<text:`api.`>-><btree:[<any:![.]><-<text:`.com`>-><nothing>]>]>
	// true
	// false
	// <prefix_suffix:[api.,.com]>
	// true
	// true
	// <row_3:[<single> <text:`at`>]>
	// true
	// true
	// false
	// <row_3:[<single:![f]> <text:`at`>]>
	// true
	// false
	// false
	// <row_3:[<list:[abc]> <text:`at`>]>
	// true
	// true
	// false
	// false
	// <row_3:[<list:![abc]> <text:`at`>]>
	// false
	// false
	// true
	// false
	// <row_3:[<range:[a,c]> <text:`at`>]>
	// true
	// true
	// false
	// false
	// <row_3:[<range:![a,c]> <text:`at`>]>
	// false
	// false
	// true
	// false
	// <indexed_any_of:[[<text:`cat`> <text:`bat`> <row_3:[<list:[fr]> <text:`at`>]>]]>
	// true
	// true
	// true
	// true
	// false
	// false
	// false
}
