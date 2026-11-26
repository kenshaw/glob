package glob

import (
	"encoding/json"
	"slices"
	"strconv"
	"testing"
)

// TestUpstreamIssues runs the cases of the issues of upstream, from
// issues_test.go of upstream at ff0540d (D7). Each case names the number of
// its issue on upstream.
func TestUpstreamIssues(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		issue  int
		pat, s string
		sep    []rune
		exp    bool
	}{
		{33, "test*201901*.12*", "test.20190101.120000", nil, true},
		{33, "test.*201901*.14*", "test.20190101.130000", nil, false},
		{33, "test.*201901*.14*", "test.20190101.140000", nil, true},
		{36, "test,pattern", "test,pattern", nil, true},
		{36, `test\,pattern`, "test,pattern", nil, true},
		{41, "hell?", "hello", nil, true},
		{41, "hell[o]", "hello", nil, true},
		{41, "ångstr?m", "ångström", nil, true},
		{41, "ångstr[ö]m", "ångström", nil, true},
		{43, "{*.,}google*", "a.google.com", nil, true},
		{43, "{*.,}google*", "agoogle.com", nil, false},
		{43, "{*.,}google*", "google.com", nil, true},
		{43, "{,*.}google*", "a.google.com", nil, true},
		{43, "{,*.}google*", "agoogle.com", nil, false},
		{43, "{,*.}google*", "google.com", nil, true},
		{50, "*yandex:*.exe:page.*", "yandex:service.exe:page.12345", nil, true},
		{50, "yandex:*.exe:page.*", "yandex:service.exe:page.12345", nil, true},
		{50, "{*yandex:*.exe:page.*,x}", "yandex:service.exe:page.12345", nil, true},
		{50, "{*yandex:*.exe:page.*}", "yandex:service.exe:page.12345", nil, true},
		{51, "a*ant", "an ant", nil, true},
		{51, "a*ant", "ant", nil, false},
		{51, "br*r", "br", nil, false},
		{51, "br*r", "brother", nil, true},
		{51, "so*so", "so so", nil, true},
		{51, "so*so", "so", nil, false},
		{51, "{a*ant,q}", "ant", nil, false},
		{51, "x{so*so}", "xso", nil, false},
		{54, "1?5helloö", "155helloö", nil, true},
		{54, "1?5ö", "155ö", nil, true},
		{54, "1?ö5hello", "15ö5hello", nil, true},
		{54, "1ö?5", "1ö55", nil, true},
		{58, "a/**/b", "a/b", []rune{'/'}, false},
		{58, "a/**/b", "a//b", []rune{'/'}, true},
		{58, "{**/,}b", "b", []rune{'/'}, true},
		{61, "start*art", "start", nil, false},
		{62, "{a,}{a,}a", "a", nil, true},
		{66, `{**/daxing,daxing,x,y}/**/*dev*.yaml`, "playground/daxing/generated/dev.yaml", nil, true},
		{66, `{**/daxing,daxing,x}/**/*dev*.yaml`, "playground/daxing/generated/dev.yaml", nil, true},
		{66, `{**/daxing,daxing}/**/*dev*.yaml`, "playground/daxing/generated/dev.yaml", nil, true},
		{66, `{**/daxing}/**/*dev*.yaml`, "playground/daxing/generated/dev.yaml", nil, true},
		{76, "/**/index.org", "/index.org", []rune{'/'}, false},
		{83, `[\-a]`, "-", nil, true},
		{83, `[\-a]`, "a", nil, true},
		{83, `[\-a]`, "_", nil, false},
		{83, `[\-]`, "-", nil, true},
	} {
		g, err := Compile(test.pat, test.sep...)
		if err != nil {
			t.Errorf("issue %d: Compile(%q): expected no error, got: %v", test.issue, test.pat, err)
			continue
		}
		if got := g.Match(test.s); got != test.exp {
			t.Errorf("issue %d: Compile(%q).Match(%q): expected %t, got: %t", test.issue, test.pat, test.s, test.exp, got)
		}
	}
}

// TestMalformed holds D8. Each pattern is not complete, and Compile returns
// an error for it.
func TestMalformed(t *testing.T) {
	t.Parallel()
	for _, pattern := range []string{
		"0{",
		"/{a{*.json{",
		"{a,b",
		"a{b",
		"[",
		"[a",
		"[!",
		"[!]",
		"[]",
		"[a-",
		`[a\`,
		`abc\`,
		`\`,
		"@(a",
		"!(a|b",
		"+(a{b)",
		"[[:alfa:]]",
		"[z-a]",
		"a\xffb",
		"[\xff]",
		"{1..100000}",
	} {
		if _, err := Compile(pattern); err == nil {
			t.Errorf("Compile(%q): expected an error, got: nil", pattern)
		}
	}
}

// TestCharacters holds D8. A character that has a meaning only inside a group
// is a character outside one, and U+FFFD and NUL are characters.
func TestCharacters(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		pattern, s string
		exp        bool
	}{
		{"a}", "a}", true},
		{"}", "}", true},
		{"a)b", "a)b", true},
		{"a|b", "a|b", true},
		{"a,b", "a,b", true},
		{"(a)", "(a)", true},
		{"+", "+", true},
		{"a+b", "a+b", true},
		{"@", "@", true},
		{"!", "!", true},
		{"!a", "!a", true},
		{"�", "�", true},
		{"[�]", "�", true},
		{"[!a]", "�", true},
		{"a\x00b", "a\x00b", true},
		{"a\x00b", "a", false},
		{"**\x00", "0", false},
		{`\@(a)`, "@(a)", true},
		{`@\(a)`, "@(a)", true},
	} {
		g, err := Compile(test.pattern)
		if err != nil {
			t.Errorf("Compile(%q): expected no error, got: %v", test.pattern, err)
			continue
		}
		if got := g.Match(test.s); got != test.exp {
			t.Errorf("Compile(%q).Match(%q): expected %t, got: %t", test.pattern, test.s, test.exp, got)
		}
	}
}

// TestOptions holds D13, D16 and D17.
func TestOptions(t *testing.T) {
	t.Parallel()
	for i, test := range []struct {
		pattern string
		opts    []Option
		s       string
		exp     bool
	}{
		{"*.go", nil, "MAIN.GO", false},
		{"*.go", []Option{WithFold()}, "MAIN.GO", true},
		{"*.GO", []Option{WithFold()}, "main.go", true},
		{"[a-c]x", []Option{WithFold()}, "BX", true},
		{"[a-c]x", []Option{WithFold()}, "dx", false},
		{"[!a-c]x", []Option{WithFold()}, "BX", false},
		{"[[:lower:]]", []Option{WithFold()}, "Q", true},
		{"[[:upper:]]", nil, "q", false},
		{"straße", []Option{WithFold()}, "STRAẞE", true},
		{"k", []Option{WithFold()}, "K", true}, // the Kelvin sign
		{"s", []Option{WithFold()}, "ſ", true},
		{"*X*", []Option{WithFold(), WithSeparators('x')}, "axbxc", false},
		{"*X*", []Option{WithFold(), WithSeparators('x')}, "aXb", true},
		{"**X**", []Option{WithFold(), WithSeparators('x')}, "aXb", true},
		{"@(foo|bar)", []Option{WithFold()}, "FOO", true},
		{"a/**/b", []Option{WithSeparators('/')}, "a/b", false},
		{"a/**/b", []Option{WithSeparators('/'), WithGlobstar()}, "a/b", true},
		{"a/**/b", []Option{WithSeparators('/'), WithGlobstar()}, "a/x/y/b", true},
		{"a/**/b", []Option{WithSeparators('/'), WithGlobstar()}, "a/xb", false},
		{"**/b", []Option{WithSeparators('/'), WithGlobstar()}, "b", true},
		{"**/b", []Option{WithSeparators('/'), WithGlobstar()}, "x/b", true},
		{"**/b", []Option{WithSeparators('/'), WithGlobstar()}, "xb", false},
		{"a**/b", []Option{WithSeparators('/'), WithGlobstar()}, "ab", false},
		{"a/**", []Option{WithSeparators('/'), WithGlobstar()}, "a/x/y", true},
		{"a/**/**/b", []Option{WithSeparators('/'), WithGlobstar()}, "a/b", true},
		{"{x/**/b,c}", []Option{WithSeparators('/'), WithGlobstar()}, "x/b", true},
		{"{**/b,c}", []Option{WithSeparators('/'), WithGlobstar()}, "b", true},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			g, err := CompileWith(test.pattern, test.opts...)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if got := g.Match(test.s); got != test.exp {
				t.Errorf("%q, %q: expected %t, got: %t", test.pattern, test.s, test.exp, got)
			}
			if got := g.MatchBytes([]byte(test.s)); got != test.exp {
				t.Errorf("%q, %q: MatchBytes: expected %t, got: %t", test.pattern, test.s, test.exp, got)
			}
		})
	}
}

func TestSeparators(t *testing.T) {
	t.Parallel()
	g := Must("*", '.', '/')
	if got, exp := g.Separators(), []rune{'.', '/'}; !slices.Equal(got, exp) {
		t.Errorf("expected %q, got: %q", string(exp), string(got))
	}
	g.Separators()[0] = 'x'
	if got := g.Separators(); got[0] != '.' {
		t.Errorf("Separators returned the slice of the glob, which a caller changed to %q", string(got))
	}
	if got := Must("*").Separators(); len(got) != 0 {
		t.Errorf("expected no separator, got: %q", string(got))
	}
}

// TestUnmarshalOptions holds D13. A glob from New keeps its options for
// UnmarshalText.
func TestUnmarshalOptions(t *testing.T) {
	t.Parallel()
	g := New(WithSeparators('.'), WithFold())
	if err := g.UnmarshalText([]byte("*.EXAMPLE.com")); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	for s, exp := range map[string]bool{
		"www.example.com":   true,
		"a.b.example.com":   false,
		"WWW.Example.COM":   true,
		"www.example.com.x": false,
	} {
		if got := g.Match(s); got != exp {
			t.Errorf("%q: expected %t, got: %t", s, exp, got)
		}
	}
	if err := g.UnmarshalText([]byte("a{")); err == nil {
		t.Errorf("expected an error, got: nil")
	}
}

// TestPath holds D14.
func TestPath(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		pattern, s string
		exp        bool
	}{
		{"a/**/b", "a/b", true},
		{"a/**/b", "a/x/y/b", true},
		{"a/*/b", "a/x/y/b", false},
		{"**/*.go", "main.go", true},
		{"**/*.go", "cmd/glob/main.go", true},
		{"*.go", "cmd/main.go", false},
		{"src/{a,b}/*.[ch]", "src/b/x.h", true},
		{"!(*.go)", "main.c", true},
		{"!(*.go)", "dir/main.c", false},
	} {
		p, err := CompilePath(test.pattern)
		if err != nil {
			t.Fatalf("CompilePath(%q): expected no error, got: %v", test.pattern, err)
		}
		if got := p.Match(test.s); got != test.exp {
			t.Errorf("CompilePath(%q).Match(%q): expected %t, got: %t", test.pattern, test.s, test.exp, got)
		}
		if got := p.MatchBytes([]byte(test.s)); got != test.exp {
			t.Errorf("CompilePath(%q).MatchBytes(%q): expected %t, got: %t", test.pattern, test.s, test.exp, got)
		}
	}
	p, err := CompilePath("a/**/b", WithGlobstar(), WithFold())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !p.Match("A/B") {
		t.Errorf("an option of CompilePath did not apply")
	}
}

// TestHost holds D14.
func TestHost(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		pattern, s string
		exp        bool
	}{
		{"*.example.com", "www.example.com", true},
		{"*.example.com", "WWW.EXAMPLE.COM", true},
		{"*.example.com", "www.example.com.", true},
		{"*.example.com.", "www.example.com", true},
		{"*.example.com", "a.b.example.com", false},
		{"**.example.com", "a.b.example.com", true},
		{"*.example.com", "example.com", false},
		{"{,*.}example.com", "example.com", true},
		{"api-[0-9].example.com", "API-7.example.com", true},
	} {
		h, err := CompileHost(test.pattern)
		if err != nil {
			t.Fatalf("CompileHost(%q): expected no error, got: %v", test.pattern, err)
		}
		if got := h.Match(test.s); got != test.exp {
			t.Errorf("CompileHost(%q).Match(%q): expected %t, got: %t", test.pattern, test.s, test.exp, got)
		}
		if got := h.MatchBytes([]byte(test.s)); got != test.exp {
			t.Errorf("CompileHost(%q).MatchBytes(%q): expected %t, got: %t", test.pattern, test.s, test.exp, got)
		}
		if got := h.String(); got != test.pattern {
			t.Errorf("expected %q, got: %q", test.pattern, got)
		}
	}
}

// TestTextFields holds D14. A struct field of type Glob, Path or Host
// unmarshals from JSON with the options of its type.
func TestTextFields(t *testing.T) {
	t.Parallel()
	var v struct {
		Glob  Glob   `json:"glob"`
		Paths []Path `json:"paths"`
		Hosts []Host `json:"hosts"`
	}
	const in = `{"glob":"a*","paths":["src/**/*.go"],"hosts":["*.Example.com"]}`
	if err := json.Unmarshal([]byte(in), &v); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	switch {
	case !v.Glob.Match("abc"):
		t.Errorf("the glob did not match")
	case !v.Paths[0].Match("src/main.go"):
		t.Errorf("the path did not match a path with no segment for **")
	case v.Paths[0].Match("src/x/main.c"):
		t.Errorf("the path matched a file that is not Go")
	case !v.Hosts[0].Match("www.example.COM"):
		t.Errorf("the host did not match without regard to case")
	case v.Hosts[0].Match("a.b.example.com"):
		t.Errorf("the host matched across a dot")
	}
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if string(out) != in {
		t.Errorf("expected %s, got: %s", in, out)
	}
	if err := json.Unmarshal([]byte(`{"paths":["a{"]}`), &v); err == nil {
		t.Errorf("expected an error, got: nil")
	}
}
