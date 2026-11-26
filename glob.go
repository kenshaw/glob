// Package glob provides glob matching.
package glob

import (
	"path/filepath"
	"slices"
	"strings"
	"unsafe"

	"github.com/kenshaw/glob/syntax"
)

// Glob matches glob patterns.
type Glob struct {
	syntax.Matcher

	pattern string
	opts    options
}

// New creates a new, empty glob with the options. [Glob.UnmarshalText]
// compiles a pattern with these options.
func New(opts ...Option) *Glob {
	g := new(Glob)
	for _, o := range opts {
		o(&g.opts)
	}
	return g
}

// Compile creates a [Glob] for given pattern and strings (if any present after
// pattern) as separators. It is the same as [CompileWith] with the option
// [WithSeparators]. The pattern syntax is:
//
//	pattern:
//	    { term }
//
//	term:
//	    `*`         matches any sequence of non-separator characters
//	    `**`        matches any sequence of characters
//	    `?`         matches any single non-separator character
//	    `[` [ `!` | `^` ] { class-item } `]`
//	                character class (must be non-empty)
//	    `{` pattern-list `}`
//	                pattern alternatives
//	    `{` lo `..` hi [ `..` step ] `}`
//	                a sequence of numbers or of letters, as in {1..10},
//	                {01..10..2} or {a..e}
//	    `?(` ext-list `)`
//	                matches zero or one of the patterns
//	    `*(` ext-list `)`
//	                matches zero or more of the patterns
//	    `+(` ext-list `)`
//	                matches one or more of the patterns
//	    `@(` ext-list `)`
//	                matches exactly one of the patterns
//	    `!(` ext-list `)`
//	                matches any sequence of non-separator characters that
//	                none of the patterns matches
//	    c           matches character c (c != `*`, `**`, `?`, `\`, `[`, `{`, `}`)
//	    `\` c       matches character c
//
//	class-item:
//	    c           matches character c (c != `\`, `-`, `]`)
//	    `\` c       matches character c
//	    lo `-` hi   matches character c for lo <= c <= hi
//	    `[:` name `:]`
//	                matches a character of the named class: alnum, alpha,
//	                ascii, blank, cntrl, digit, graph, lower, print, punct,
//	                space, upper, word or xdigit
//	    `[=` c `=]` matches character c
//	    `[.` c `.]` matches character c
//
//	pattern-list:
//	    pattern { `,` pattern }
//	                comma-separated (without spaces) patterns
//
//	ext-list:
//	    pattern { `|` pattern }
//	                patterns separated by a vertical bar
//
// A `]` that is the first character of a class, after an optional `!` or
// `^`, is a character of the class, and so is a `-` that is first or last.
func Compile(pattern string, separators ...rune) (*Glob, error) {
	return CompileWith(pattern, WithSeparators(separators...))
}

// CompileWith creates a [Glob] for the pattern with the options. [Compile]
// describes the syntax of the pattern.
func CompileWith(pattern string, opts ...Option) (*Glob, error) {
	g := New(opts...)
	if err := g.compile(pattern); err != nil {
		return nil, err
	}
	return g, nil
}

// Must is the same as Compile, except that if Compile returns error, this will
// panic.
func Must(pattern string, separators ...rune) *Glob {
	g, err := Compile(pattern, separators...)
	if err != nil {
		panic(err)
	}
	return g
}

// compile compiles the pattern with the options of the glob.
func (g *Glob) compile(pattern string) error {
	m, err := g.opts.compile(pattern)
	if err != nil {
		return err
	}
	g.Matcher, g.pattern = m, pattern
	return nil
}

// Match reports whether the glob matches s.
func (g *Glob) Match(s string) bool {
	if g.opts.fold {
		s = syntax.FoldString(s)
	}
	return g.Matcher.Match(s)
}

// MatchBytes reports whether the glob matches b. It does not copy b.
func (g *Glob) MatchBytes(b []byte) bool {
	// No matcher keeps the string after Match returns.
	return g.Match(unsafe.String(unsafe.SliceData(b), len(b)))
}

// Separators returns the separators of the glob.
func (g *Glob) Separators() []rune {
	return slices.Clone(g.opts.sep)
}

// UnmarshalText satisfies the [encoding.TextUnmarshaler] interface. It
// compiles the text with the options of the glob. See [New].
func (g *Glob) UnmarshalText(buf []byte) error {
	return g.compile(string(buf))
}

// MarshalText satisfies the [encoding.TextMarshaler] interface. It has a value
// receiver, so a struct field of type Glob marshals as its pattern.
func (g Glob) MarshalText() ([]byte, error) {
	return []byte(g.pattern), nil
}

// String satisfies the [fmt.Stringer] interface.
func (g Glob) String() string {
	return g.pattern
}

// Option is an option of [CompileWith] and [New].
type Option func(*options)

// WithSeparators is an option to set the separators. A `*` or a `?` does not
// match a separator, and a `**` does.
func WithSeparators(separators ...rune) Option {
	return func(o *options) {
		o.sep = slices.Clone(separators)
	}
}

// WithFold is an option to match without regard to case, as
// [strings.EqualFold] compares.
func WithFold() Option {
	return func(o *options) {
		o.fold = true
	}
}

// WithGlobstar is an option to make a `**` that is a whole segment also
// match no segment at all, as the globstar of a shell does. With the
// separator `/`, the pattern a/**/b then matches a/b as well as a/x/b.
func WithGlobstar() Option {
	return func(o *options) {
		o.globstar = true
	}
}

// options are the options of a glob.
type options struct {
	sep      []rune
	fold     bool
	globstar bool
}

// compile compiles the pattern with the options.
func (o options) compile(pattern string) (syntax.Matcher, error) {
	tree, err := syntax.Parse(syntax.NewLexer(pattern))
	if err != nil {
		return nil, err
	}
	sep := o.sep
	if o.globstar {
		tree = syntax.Globstar(tree, sep)
	}
	if o.fold {
		tree = syntax.Fold(tree)
		sep = []rune(syntax.FoldString(string(sep)))
	}
	return tree.Match(sep)
}

// Path is a glob that matches slash-separated paths. Its separator is `/`, and
// a `**` that is a whole segment matches no segment too. See [WithGlobstar].
type Path struct {
	Glob
}

// CompilePath creates a [Path] for the pattern. The options follow the options
// of a path.
func CompilePath(pattern string, opts ...Option) (*Path, error) {
	p := new(Path)
	p.opts = pathOptions(opts...)
	if err := p.compile(pattern); err != nil {
		return nil, err
	}
	return p, nil
}

// pathOptions returns the options of a path, then opts.
func pathOptions(opts ...Option) options {
	o := options{sep: []rune{'/'}, globstar: true}
	for _, f := range opts {
		f(&o)
	}
	return o
}

// Match reports whether the path matches name. It changes each separator of
// the operating system in name to a slash first. See [filepath.ToSlash].
func (p *Path) Match(name string) bool {
	return p.Glob.Match(filepath.ToSlash(name))
}

// MatchBytes reports whether the path matches b.
func (p *Path) MatchBytes(b []byte) bool {
	return p.Match(string(b))
}

// UnmarshalText satisfies the [encoding.TextUnmarshaler] interface.
func (p *Path) UnmarshalText(buf []byte) error {
	p.opts = pathOptions()
	return p.compile(string(buf))
}

// Host is a glob that matches host names. Its separator is `.`, and it
// matches without regard to case. It ignores one dot at the end of a name,
// so the name example.com. matches as example.com does.
type Host struct {
	Glob
}

// CompileHost creates a [Host] for the pattern. The options follow the options
// of a host.
func CompileHost(pattern string, opts ...Option) (*Host, error) {
	h := new(Host)
	h.opts = hostOptions(opts...)
	if err := h.compile(strings.TrimSuffix(pattern, ".")); err != nil {
		return nil, err
	}
	h.pattern = pattern
	return h, nil
}

// hostOptions returns the options of a host, then opts.
func hostOptions(opts ...Option) options {
	o := options{sep: []rune{'.'}, fold: true}
	for _, f := range opts {
		f(&o)
	}
	return o
}

// Match reports whether the host matches name.
func (h *Host) Match(name string) bool {
	return h.Glob.Match(strings.TrimSuffix(name, "."))
}

// MatchBytes reports whether the host matches b.
func (h *Host) MatchBytes(b []byte) bool {
	return h.Match(string(b))
}

// UnmarshalText satisfies the [encoding.TextUnmarshaler] interface.
func (h *Host) UnmarshalText(buf []byte) error {
	h.opts = hostOptions()
	if err := h.compile(strings.TrimSuffix(string(buf), ".")); err != nil {
		return err
	}
	h.pattern = string(buf)
	return nil
}

// Quote returns a string that escapes each special character of the syntax
// in s. For example, Quote(`{foo*}`) returns `\{foo\*\}`. A glob compiled
// from the result matches s and nothing else.
func Quote(s string) string {
	b := make([]byte, 0, 2*len(s))
	// a byte loop is correct because all meta characters are ASCII
	for i := range len(s) {
		if syntax.IsSpecial(s[i]) {
			b = append(b, '\\')
		}
		b = append(b, s[i])
	}
	return string(b)
}
