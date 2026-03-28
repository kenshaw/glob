package glob

import (
	"strconv"
	"testing"
)

func TestCompile(t *testing.T) {
	for i, test := range []struct {
		s   string
		v   string
		sep rune
		exp bool
	}{
		{
			``,
			``,
			0, true,
		},
		{
			`b`,
			``,
			0, false,
		},
		{
			`åä`,
			`*ä`,
			0, true,
		},
		{
			`abc`,
			`abc`,
			0, true,
		},
		{
			`abc`,
			`a*c`,
			0, true,
		},
		{
			`a12345c`,
			`a*c`,
			0, true,
		},
		{
			`a1c`,
			`a?c`,
			0, true,
		},
		{
			`a.b`,
			`a.b`,
			'.', true,
		},
		{
			`a.b`,
			`a.*`,
			'.', true,
		},
		{
			`a.b.c`,
			`a.**`,
			'.', true,
		},
		{
			`a.b.c`,
			`a.?.c`,
			'.', true,
		},
		{
			`a.b.c`,
			`a.?.?`,
			'.', true,
		},
		{
			`cat`,
			`?at`,
			0, true,
		},
		{
			`fat`,
			`?at`,
			0, true,
		},
		{
			`abc`,
			`*`,
			0, true,
		},
		{
			`*`,
			`\*`,
			0, true,
		},
		{
			`a.b.c`,
			`**`,
			'.', true,
		},
		{
			`at`,
			`?at`,
			0, false,
		},
		{
			`fat`,
			`?at`,
			'f', false,
		},
		{
			`a.b.c`,
			`a.*`,
			'.', false,
		},
		{
			`a.bb.c`,
			`a.?.c`,
			'.', false,
		},
		{
			`a.b.c`,
			`*`,
			'.', false,
		},
		{
			`this is a test`,
			`*test`,
			0, true,
		},
		{
			`this is a test`,
			`this*`,
			0, true,
		},
		{
			`this is a test`,
			`*is *`,
			0, true,
		},
		{
			`this is a test`,
			`*is*a*`,
			0, true,
		},
		{
			`this is a test`,
			`**test**`,
			0, true,
		},
		{
			`this is a test`,
			`**is**a***test*`,
			0, true,
		},
		{
			`this is a test`,
			`*is`,
			0, false,
		},
		{
			`this is a test`,
			`*no*`,
			0, false,
		},
		{
			`this is a test3`,
			`[!a]*`,
			0, true,
		},
		{
			`this is a test3`,
			`[!t]*`,
			0, false,
		},
		{
			`abcabc`,
			`*abc`,
			0, true,
		},
		{
			`abcabc`,
			`**abc`,
			0, true,
		},
		{
			`abc`,
			`???`,
			0, true,
		},
		{
			`abc`,
			`?*?`,
			0, true,
		},
		{
			`ac`,
			`?*?`,
			0, true,
		},
		{
			`stagnation`,
			`sta`,
			0, false,
		},
		{
			`stagnation`,
			`sta*`,
			0, true,
		},
		{
			`stagnation`,
			`STA*`,
			0, false,
		},
		{
			`stagnation`,
			`sta?`,
			0, false,
		},
		{
			`stagnation`,
			`sta?n`,
			0, false,
		},
		{
			`my cat has very bright eyes`,
			`* ?at * eyes`,
			0, true,
		},
		{
			`defghi`,
			`{abc,def}ghi`,
			0, true,
		},
		{
			`abcda`,
			`{abc,abcd}a`,
			0, true,
		},
		{
			`abc`,
			`{a,ab}{bc,f}`,
			0, true,
		},
		{
			`ab`,
			`{*,**}{a,b}`,
			0, true,
		},
		{
			`ac`,
			`{*,**}{a,b}`,
			0, false,
		},
		{
			`/rate`,
			`/{rate,[a-z][a-z][a-z]}*`,
			0, true,
		},
		{
			`/rate`,
			`/{rate,[0-9][0-9][0-9]}*`,
			0, true,
		},
		{
			`/usd`,
			`/{rate,[a-z][a-z][a-z]}*`,
			0, true,
		},
		{
			`www.google.com`,
			`{*.google.*,*.yandex.*}`,
			'.', true,
		},
		{
			`www.yandex.com`,
			`{*.google.*,*.yandex.*}`,
			'.', true,
		},
		{
			`yandex.com`,
			`{*.google.*,*.yandex.*}`,
			'.', false,
		},
		{
			`google.com`,
			`{*.google.*,*.yandex.*}`,
			'.', false,
		},
		{
			`www.google.com`,
			`{*.google.*,yandex.*}`,
			'.', true,
		},
		{
			`yandex.com`,
			`{*.google.*,yandex.*}`,
			'.', true,
		},
		{
			`www.yandex.com`,
			`{*.google.*,yandex.*}`,
			'.', false,
		},
		{
			`google.com`,
			`{*.google.*,yandex.*}`,
			'.', false,
		},
		{
			`https://www.example.com`,
			`*//{,*.}example.com`,
			0, true,
		},
		{
			`http://example.com`,
			`*//{,*.}example.com`,
			0, true,
		},
		{
			`http://example.com.net`,
			`*//{,*.}example.com`,
			0, false,
		},
		{
			`abc`,
			`{a*,b}c`,
			'.', true,
		},
		{fixture_all_match, pattern_all, 0, true},
		{fixture_all_mismatch, pattern_all, 0, false},
		{fixture_plain_match, pattern_plain, 0, true},
		{fixture_plain_mismatch, pattern_plain, 0, false},
		{fixture_multiple_match, pattern_multiple, 0, true},
		{fixture_multiple_mismatch, pattern_multiple, 0, false},
		{fixture_alternatives_match, pattern_alternatives, 0, true},
		{fixture_alternatives_mismatch, pattern_alternatives, 0, false},
		{fixture_alternatives_suffix_first_match, pattern_alternatives_suffix, 0, true},
		{fixture_alternatives_suffix_first_mismatch, pattern_alternatives_suffix, 0, false},
		{fixture_alternatives_suffix_second, pattern_alternatives_suffix, 0, true},
		{fixture_alternatives_combine_hard, pattern_alternatives_combine_hard, 0, true},
		{fixture_alternatives_combine_lite, pattern_alternatives_combine_lite, 0, true},
		{fixture_prefix_suffix_match, pattern_prefix, 0, true},
		{fixture_prefix_suffix_mismatch, pattern_prefix, 0, false},
		{fixture_prefix_suffix_match, pattern_suffix, 0, true},
		{fixture_prefix_suffix_mismatch, pattern_suffix, 0, false},
		{fixture_prefix_suffix_match, pattern_prefix_suffix, 0, true},
		{fixture_prefix_suffix_mismatch, pattern_prefix_suffix, 0, false},
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
