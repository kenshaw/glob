package glob

import (
	"strconv"
	"testing"
)

func TestCompile(t *testing.T) {
	for i, test := range []globTest{
		g(true, "* ?at * eyes", "my cat has very bright eyes"),
		g(true, "", ""),
		g(false, "", "b"),
		g(true, "*ä", "åä"),
		g(true, "abc", "abc"),
		g(true, "a*c", "abc"),
		g(true, "a*c", "a12345c"),
		g(true, "a?c", "a1c"),
		g(true, "a.b", "a.b", '.'),
		g(true, "a.*", "a.b", '.'),
		g(true, "a.**", "a.b.c", '.'),
		g(true, "a.?.c", "a.b.c", '.'),
		g(true, "a.?.?", "a.b.c", '.'),
		g(true, "?at", "cat"),
		g(true, "?at", "fat"),
		g(true, "*", "abc"),
		g(true, `\*`, "*"),
		g(true, "**", "a.b.c", '.'),
		g(false, "?at", "at"),
		g(false, "?at", "fat", 'f'),
		g(false, "a.*", "a.b.c", '.'),
		g(false, "a.?.c", "a.bb.c", '.'),
		g(false, "*", "a.b.c", '.'),
		g(true, "*test", "this is a test"),
		g(true, "this*", "this is a test"),
		g(true, "*is *", "this is a test"),
		g(true, "*is*a*", "this is a test"),
		g(true, "**test**", "this is a test"),
		g(true, "**is**a***test*", "this is a test"),
		g(false, "*is", "this is a test"),
		g(false, "*no*", "this is a test"),
		g(true, "[!a]*", "this is a test3"),
		g(true, "*abc", "abcabc"),
		g(true, "**abc", "abcabc"),
		g(true, "???", "abc"),
		g(true, "?*?", "abc"),
		g(true, "?*?", "ac"),
		g(false, "sta", "stagnation"),
		g(true, "sta*", "stagnation"),
		g(false, "sta?", "stagnation"),
		g(false, "sta?n", "stagnation"),
		g(true, "{abc,def}ghi", "defghi"),
		g(true, "{abc,abcd}a", "abcda"),
		g(true, "{a,ab}{bc,f}", "abc"),
		g(true, "{*,**}{a,b}", "ab"),
		g(false, "{*,**}{a,b}", "ac"),
		g(true, "/{rate,[a-z][a-z][a-z]}*", "/rate"),
		g(true, "/{rate,[0-9][0-9][0-9]}*", "/rate"),
		g(true, "/{rate,[a-z][a-z][a-z]}*", "/usd"),
		g(true, "{*.google.*,*.yandex.*}", "www.google.com", '.'),
		g(true, "{*.google.*,*.yandex.*}", "www.yandex.com", '.'),
		g(false, "{*.google.*,*.yandex.*}", "yandex.com", '.'),
		g(false, "{*.google.*,*.yandex.*}", "google.com", '.'),
		g(true, "{*.google.*,yandex.*}", "www.google.com", '.'),
		g(true, "{*.google.*,yandex.*}", "yandex.com", '.'),
		g(false, "{*.google.*,yandex.*}", "www.yandex.com", '.'),
		g(false, "{*.google.*,yandex.*}", "google.com", '.'),
		g(true, "*//{,*.}example.com", "https://www.example.com"),
		g(true, "*//{,*.}example.com", "http://example.com"),
		g(false, "*//{,*.}example.com", "http://example.com.net"),
		g(true, "{a*,b}c", "abc", '.'),
		g(true, pattern_all, fixture_all_match),
		g(false, pattern_all, fixture_all_mismatch),
		g(true, pattern_plain, fixture_plain_match),
		g(false, pattern_plain, fixture_plain_mismatch),
		g(true, pattern_multiple, fixture_multiple_match),
		g(false, pattern_multiple, fixture_multiple_mismatch),
		g(true, pattern_alternatives, fixture_alternatives_match),
		g(false, pattern_alternatives, fixture_alternatives_mismatch),
		g(true, pattern_alternatives_suffix, fixture_alternatives_suffix_first_match),
		g(false, pattern_alternatives_suffix, fixture_alternatives_suffix_first_mismatch),
		g(true, pattern_alternatives_suffix, fixture_alternatives_suffix_second),
		g(true, pattern_alternatives_combine_hard, fixture_alternatives_combine_hard),
		g(true, pattern_alternatives_combine_lite, fixture_alternatives_combine_lite),
		g(true, pattern_prefix, fixture_prefix_suffix_match),
		g(false, pattern_prefix, fixture_prefix_suffix_mismatch),
		g(true, pattern_suffix, fixture_prefix_suffix_match),
		g(false, pattern_suffix, fixture_prefix_suffix_mismatch),
		g(true, pattern_prefix_suffix, fixture_prefix_suffix_match),
		g(false, pattern_prefix_suffix, fixture_prefix_suffix_mismatch),
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Logf("%q (%q) :: %q -> %t", test.s, string(test.sep), test.m, test.exp)
			g, err := Compile(test.s, test.sep...)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if b := g.Match(test.m); b != test.exp {
				t.Errorf("pattern %q matching %q should be %v but got %v\n%s", test.s, test.m, test.exp, b, g)
			}
		})
	}
}

func TestCompileSeparators(t *testing.T) {
	for i, test := range []globTest{
		gc("{*,**,?}", '.'),
		gc("{*.google.*,yandex.*}", '.'),
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := Compile(test.s, test.sep...)
			if err != nil {
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

type globTest struct {
	s   string
	m   string
	exp bool
	sep []rune
}

func g(exp bool, s, m string, sep ...rune) globTest {
	return globTest{
		s:   s,
		m:   m,
		exp: exp,
		sep: sep,
	}
}

func gc(s string, del ...rune) globTest {
	return globTest{s: s, sep: del}
}

const (
	pattern_all                                = "[a-z][!a-x]*cat*[h][!b]*eyes*"
	regexp_all                                 = `^[a-z][^a-x].*cat.*[h][^b].*eyes.*$`
	fixture_all_match                          = "my cat has very bright eyes"
	fixture_all_mismatch                       = "my dog has very bright eyes"
	pattern_plain                              = "google.com"
	regexp_plain                               = `^google\.com$`
	fixture_plain_match                        = "google.com"
	fixture_plain_mismatch                     = "kenshaw.com"
	pattern_multiple                           = "https://*.google.*"
	regexp_multiple                            = `^https:\/\/.*\.google\..*$`
	fixture_multiple_match                     = "https://account.google.com"
	fixture_multiple_mismatch                  = "https://google.com"
	pattern_alternatives                       = "{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}"
	regexp_alternatives                        = `^(https:\/\/.*\.google\..*|.*yandex\..*|.*yahoo\..*|.*mail\.ru)$`
	fixture_alternatives_match                 = "http://yahoo.com"
	fixture_alternatives_mismatch              = "http://google.com"
	pattern_alternatives_suffix                = "{https://*kenshaw.com,http://exclude.kenshaw.com}"
	regexp_alternatives_suffix                 = `^(https:\/\/.*kenshaw\.com|http://exclude.kenshaw.com)$`
	fixture_alternatives_suffix_first_match    = "https://safe.kenshaw.com"
	fixture_alternatives_suffix_first_mismatch = "http://safe.kenshaw.com"
	fixture_alternatives_suffix_second         = "http://exclude.kenshaw.com"
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
