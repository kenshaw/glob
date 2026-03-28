package glob

import (
	"strconv"
	"testing"
)

func TestCompile(t *testing.T) {
	for i, test := range []struct {
		v   string
		s   string
		sep rune
		exp bool
	}{
		{
			``,
			``,
			0, true,
		},
		{
			``,
			`b`,
			0, false,
		},
		{
			`*ä`,
			`åä`,
			0, true,
		},
		{
			`abc`,
			`abc`,
			0, true,
		},
		{
			`a*c`,
			`abc`,
			0, true,
		},
		{
			`a*c`,
			`a12345c`,
			0, true,
		},
		{
			`a?c`,
			`a1c`,
			0, true,
		},
		{
			`a.b`,
			`a.b`,
			'.', true,
		},
		{
			`a.*`,
			`a.b`,
			'.', true,
		},
		{
			`a.**`,
			`a.b.c`,
			'.', true,
		},
		{
			`a.?.c`,
			`a.b.c`,
			'.', true,
		},
		{
			`a.?.?`,
			`a.b.c`,
			'.', true,
		},
		{
			`?at`,
			`cat`,
			0, true,
		},
		{
			`?at`,
			`fat`,
			0, true,
		},
		{
			`*`,
			`abc`,
			0, true,
		},
		{
			`\*`,
			`*`,
			0, true,
		},
		{
			`**`,
			`a.b.c`,
			'.', true,
		},
		{
			`?at`,
			`at`,
			0, false,
		},
		{
			`?at`,
			`fat`,
			'f', false,
		},
		{
			`a.*`,
			`a.b.c`,
			'.', false,
		},
		{
			`a.?.c`,
			`a.bb.c`,
			'.', false,
		},
		{
			`*`,
			`a.b.c`,
			'.', false,
		},
		{
			`*test`,
			`this is a test`,
			0, true,
		},
		{
			`this*`,
			`this is a test`,
			0, true,
		},
		{
			`*is *`,
			`this is a test`,
			0, true,
		},
		{
			`*is*a*`,
			`this is a test`,
			0, true,
		},
		{
			`**test**`,
			`this is a test`,
			0, true,
		},
		{
			`**is**a***test*`,
			`this is a test`,
			0, true,
		},
		{
			`*is`,
			`this is a test`,
			0, false,
		},
		{
			`*no*`,
			`this is a test`,
			0, false,
		},
		{
			`[!a]*`,
			`this is a test3`,
			0, true,
		},
		{
			`[!t]*`,
			`this is a test3`,
			0, false,
		},
		{
			`*abc`,
			`abcabc`,
			0, true,
		},
		{
			`**abc`,
			`abcabc`,
			0, true,
		},
		{
			`???`,
			`abc`,
			0, true,
		},
		{
			`?*?`,
			`abc`,
			0, true,
		},
		{
			`?*?`,
			`ac`,
			0, true,
		},
		{
			`sta`,
			`stagnation`,
			0, false,
		},
		{
			`sta*`,
			`stagnation`,
			0, true,
		},
		{
			`STA*`,
			`stagnation`,
			0, false,
		},
		{
			`sta?`,
			`stagnation`,
			0, false,
		},
		{
			`sta?n`,
			`stagnation`,
			0, false,
		},
		{
			`* ?at * eyes`,
			`my cat has very bright eyes`,
			0, true,
		},
		{
			`{abc,def}ghi`,
			`defghi`,
			0, true,
		},
		{
			`{abc,abcd}a`,
			`abcda`,
			0, true,
		},
		{
			`{a,ab}{bc,f}`,
			`abc`,
			0, true,
		},
		{
			`{*,**}{a,b}`,
			`ab`,
			0, true,
		},
		{
			`{*,**}{a,b}`,
			`ac`,
			0, false,
		},
		{
			`/{rate,[a-z][a-z][a-z]}*`,
			`/rate`,
			0, true,
		},
		{
			`/{rate,[0-9][0-9][0-9]}*`,
			`/rate`,
			0, true,
		},
		{
			`/{rate,[a-z][a-z][a-z]}*`,
			`/usd`,
			0, true,
		},
		{
			`{*.google.*,*.yandex.*}`,
			`www.google.com`,
			'.', true,
		},
		{
			`{*.google.*,*.yandex.*}`,
			`www.yandex.com`,
			'.', true,
		},
		{
			`{*.google.*,*.yandex.*}`,
			`yandex.com`,
			'.', false,
		},
		{
			`{*.google.*,*.yandex.*}`,
			`google.com`,
			'.', false,
		},
		{
			`{*.google.*,yandex.*}`,
			`www.google.com`,
			'.', true,
		},
		{
			`{*.google.*,yandex.*}`,
			`yandex.com`,
			'.', true,
		},
		{
			`{*.google.*,yandex.*}`,
			`www.yandex.com`,
			'.', false,
		},
		{
			`{*.google.*,yandex.*}`,
			`google.com`,
			'.', false,
		},
		{
			`*//{,*.}example.com`,
			`https://www.example.com`,
			0, true,
		},
		{
			`*//{,*.}example.com`,
			`http://example.com`,
			0, true,
		},
		{
			`*//{,*.}example.com`,
			`http://example.com.net`,
			0, false,
		},
		{
			`{a*,b}c`,
			`abc`,
			'.', true,
		},
		{pattern_all, fixture_all_match, 0, true},
		{pattern_all, fixture_all_mismatch, 0, false},
		{pattern_plain, fixture_plain_match, 0, true},
		{pattern_plain, fixture_plain_mismatch, 0, false},
		{pattern_multiple, fixture_multiple_match, 0, true},
		{pattern_multiple, fixture_multiple_mismatch, 0, false},
		{pattern_alternatives, fixture_alternatives_match, 0, true},
		{pattern_alternatives, fixture_alternatives_mismatch, 0, false},
		{pattern_alternatives_suffix, fixture_alternatives_suffix_first_match, 0, true},
		{pattern_alternatives_suffix, fixture_alternatives_suffix_first_mismatch, 0, false},
		{pattern_alternatives_suffix, fixture_alternatives_suffix_second, 0, true},
		{pattern_alternatives_combine_hard, fixture_alternatives_combine_hard, 0, true},
		{pattern_alternatives_combine_lite, fixture_alternatives_combine_lite, 0, true},
		{pattern_prefix, fixture_prefix_suffix_match, 0, true},
		{pattern_prefix, fixture_prefix_suffix_mismatch, 0, false},
		{pattern_suffix, fixture_prefix_suffix_match, 0, true},
		{pattern_suffix, fixture_prefix_suffix_mismatch, 0, false},
		{pattern_prefix_suffix, fixture_prefix_suffix_match, 0, true},
		{pattern_prefix_suffix, fixture_prefix_suffix_mismatch, 0, false},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Logf("%q (%q) :: %q -> %t", test.v, string(test.sep), test.s, test.exp)
			var sep []rune
			if test.sep != 0 {
				sep = append(sep, test.sep)
			}
			g1, err := Compile(test.v, sep...)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if b := g1.Match(test.s); b != test.exp {
				t.Errorf("expected %t, got: %t", test.exp, b)
			}
			if test.sep != 0 {
				return
			}
			g2 := New()
			if err := g2.UnmarshalText([]byte(test.v)); err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if b := g2.Match(test.s); b != test.exp {
				t.Errorf("expected %t, got: %t", test.exp, b)
			}
		})
	}
}

func TestCompileSeparators(t *testing.T) {
	for i, test := range []struct {
		s   string
		sep rune
	}{
		{"{*,**,?}", '.'},
		{"{*.google.*,yandex.*}", '.'},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if _, err := Compile(test.s, test.sep); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestQuote(t *testing.T) {
	for i, test := range []struct {
		s, exp string
	}{
		{`[foo*]`, `\[foo\*\]`},
		{`{foo*}`, `\{foo\*\}`},
		{`*?\[]{}`, `\*\?\\\[\]\{\}`},
		{`some text and *?\[]{}`, `some text and \*\?\\\[\]\{\}`},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Logf("%q -> %q", test.s, test.exp)
			s := Quote(test.s)
			if s != test.exp {
				t.Errorf("QuoteMeta(%q) = %q; want %q", test.s, s, test.exp)
			}
			if _, err := Compile(s); err != nil {
				t.Errorf("_, err := Compile(QuoteMeta(%q) = %q); err = %q", test.s, s, err)
			}
		})
	}
}

const (
	pattern_all                                = "[a-z][!a-x]*cat*[h][!b]*eyes*"
	regexp_all                                 = `^[a-z][^a-x].*cat.*[h][^b].*eyes.*$`
	fixture_all_match                          = "my cat has very bright eyes"
	fixture_all_mismatch                       = "my dog has very bright eyes"
	pattern_plain                              = "google.com"
	regexp_plain                               = `^google\.com$`
	fixture_plain_match                        = "google.com"
	fixture_plain_mismatch                     = "example.com"
	pattern_multiple                           = "https://*.google.*"
	regexp_multiple                            = `^https:\/\/.*\.google\..*$`
	fixture_multiple_match                     = "https://account.google.com"
	fixture_multiple_mismatch                  = "https://google.com"
	pattern_alternatives                       = "{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}"
	regexp_alternatives                        = `^(https:\/\/.*\.google\..*|.*yandex\..*|.*yahoo\..*|.*mail\.ru)$`
	fixture_alternatives_match                 = "http://yahoo.com"
	fixture_alternatives_mismatch              = "http://google.com"
	pattern_alternatives_suffix                = "{https://*example.com,http://exclude.example.com}"
	regexp_alternatives_suffix                 = `^(https:\/\/.*example\.com|http://exclude.example.com)$`
	fixture_alternatives_suffix_first_match    = "https://safe.example.com"
	fixture_alternatives_suffix_first_mismatch = "http://safe.example.com"
	fixture_alternatives_suffix_second         = "http://exclude.example.com"
	pattern_prefix                             = "abc*"
	regexp_prefix                              = `^abc.*$`
	pattern_suffix                             = "*def"
	regexp_suffix                              = `^.*def$`
	pattern_prefix_suffix                      = "ab*ef"
	regexp_prefix_suffix                       = `^ab.*ef$`
	fixture_prefix_suffix_match                = "abcdef"
	fixture_prefix_suffix_mismatch             = "af"
	pattern_alternatives_combine_lite          = "{abc*def,abc?def,abc[zte]def}"
	regexp_alternatives_combine_lite           = `^(abc.*def|abc.def|abc[zte]def)$`
	fixture_alternatives_combine_lite          = "abczdef"
	pattern_alternatives_combine_hard          = "{abc*[a-c]def,abc?[d-g]def,abc[zte]?def}"
	regexp_alternatives_combine_hard           = `^(abc.*[a-c]def|abc.[d-g]def|abc[zte].def)$`
	fixture_alternatives_combine_hard          = "abczqdef"
)
