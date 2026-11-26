# The syntax of a pattern

This document is the survey of glob syntax that chose the syntax of the fork,
and the meaning of each part of a pattern, with its edge cases. D19 adopts it.
`testdata/vectors.tsv` holds the test vectors of the survey, and `TestVectors`
runs them.

The survey was made on 2026-11-26. It read the tests and the documents of
these implementations, and ran several of them:

- the fork, at `01e2323`, before the work of D7 to D19,
- upstream, `github.com/gobwas/glob`, at `ff0540d`,
- `path.Match` of the Go standard library,
- `github.com/bmatcuk/doublestar` version 4,
- bash 5.3, with `[[ s == p ]]` in the locale `C.UTF-8`, and its brace
  expansion,
- `fnmatch` of glibc, the POSIX function,
- minimatch, picomatch and micromatch, the matchers of npm,
- `fnmatch` and `glob` of Python,
- `globset` of Rust,
- gitignore, through `t3070-wildmatch.sh` of git,
- editorconfig, through `editorconfig-core-test`.

## The test vectors

`testdata/vectors.tsv` has these columns, separated by a tab:

| Column | What it holds |
| --- | --- |
| `feature` | the part of the syntax, and the mode that the vector needs |
| `pattern` | the pattern |
| `input` | the string to match |
| `separators` | each rune of the field is a separator. An empty field means no separator |
| `expected` | `true`, `false`, or `error` when the pattern must not compile |
| `source` | the test suite that the pattern and the input came from, and how the expected value was computed |

A field that starts with `"` is a quoted Go string. It holds a tab, a new line
or a byte that is not valid UTF-8.

The survey computed each expected value again. It did not copy it. A small
reference matcher gave the value, and the survey compared it with bash on
each vector with no separator, and with `fnmatch` and `FNM_PATHNAME` on each
vector with the separator `/`. Bash gave the values of the brace sequences,
and upstream with its option `Globstar` gave the values of the globstar rows.
Each difference that is left is in the table under Where implementations
disagree.

The vectors come from the tests of picomatch (673), the survey itself (625),
upstream (251), git (246), Rust `globset` (172), doublestar (160),
editorconfig (129), the Go standard library (56) and Python (25).

The feature names the mode:

| Feature | Mode | `TestVectors` |
| --- | --- | --- |
| `literal`, `star`, `qmark`, `separator-*`, `super`, `class`, `class-edge`, `class-separator`, `posix-class`, `equiv-collating`, `escape`, `alt`, `alt-edge`, `brace-range`, `error`, `unicode-*` | no option | runs it |
| `opt-extglob` | extended groups, which are on (D10) | runs it |
| `opt-globstar` | `WithGlobstar` | runs it |
| `opt-nocase` | `WithFold` | runs it |
| `opt-period`, `opt-leadingdir`, `opt-noescape`, `opt-negate`, `opt-matchbase` | options that the fork does not have | skips it |
| `dialect-editorconfig`, `dialect-gitignore`, `dialect-python-fnmatch` | other dialects | skips it |

## Which implementation has what

Y means yes, N means no, opt means behind an option, lit means that the
pattern is text, and err means an error. The column "fork" is the fork after
D7 to D19.

| Feature | fork | upstream | Go `path` | doublestar | bash | `fnmatch` | minimatch | picomatch | Python | `globset` | gitignore | editorconfig |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `*` stops at a separator | Y | Y | Y | Y | Y, in paths | `FNM_PATHNAME` | Y | Y | `glob` only | opt | Y | Y |
| `**` crosses separators anywhere | Y | Y | N | in a segment | N | N | in a segment | in a segment | in a segment | in a segment | in a segment | Y |
| `**/` matches no segment too | opt | opt | N | Y | opt | N | Y | Y | opt | Y | Y | Y |
| ranges and characters mixed, `[a-cx-z_]` | Y | err | Y | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `[!...]` | Y | Y | N | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `[^...]` | Y | N | Y | Y | Y | Y | Y | Y | N | Y | Y | N |
| `]` first, as in `[]a]` | Y | err | err | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `-` first or last | Y | err | err | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| `[z-a]` | err | err | no match | no match | no match | no match | no match | no match | no match | err | no match | |
| `[[:alpha:]]` | Y | N | N | N | Y | Y | Y | Y | N | Y | Y | N |
| `[[=a=]]`, `[[.a.]]` | one rune | N | N | N | Y | Y | N | N | N | N | N | N |
| a trailing `\` | err | err | err | err | lit | no match | lit | lit | lit | err | no match | |
| `{a,b}` | Y | Y | N | Y | Y | N | Y | Y | N | Y | N | Y |
| `{a}` and `{}` | a group, and empty | a group, and empty | | a group, and empty | lit | | lit | lit | | a group, and empty | | lit |
| an unclosed `{` | err | err | | err | lit | | lit | lit | | err | | lit |
| `{1..10}`, `{01..10..2}` | Y | N | N | N | Y | N | Y | micromatch | N | N | N | numbers only |
| `{a..e}` | Y | N | N | N | Y | N | Y | micromatch | N | N | N | N |
| `?()`, `*()`, `+()`, `@()`, `!()` | Y | N | N | N | opt | opt | Y | Y | N | N | N | N |
| without regard to case | opt | N | N | opt | opt | opt | opt | opt | Windows | opt | opt | N |
| a leading dot is not matched by `*` | N | N | N | opt | Y | `FNM_PERIOD` | Y | Y | `glob` only | N | N | N |
| negation of a whole pattern, `!p` | N | N | N | N | N | N | Y | Y | N | N | Y | N |
| a pattern matches the base name | N | N | N | N | N | N | opt | opt | N | N | Y | Y |

zsh and ksh93 have more syntax, which no other implementation shares: zsh has
`^p`, `p1~p2`, `#` and `##`, `<1-10>`, and flags such as `(#i)`. ksh93 has
`{n}(p)`, `&(p1&p2)` and back references. The fork leaves them out.

## The meaning of each part

The fork follows bash and `fnmatch`, with the separators of upstream. A
separator stops `?`, `*` and `!(...)`. It does not stop a character, a class or
a group that holds the separator.

### Characters, `?`, `*` and `**`

1. A pattern matches the whole input. It matches runes of UTF-8, not bytes.
2. `?` matches one rune that is not a separator. It never matches the empty
   string.
3. `*` matches any run of runes that holds no separator, and the empty run.
4. Two or more stars match any run of runes, separators too. `***` is `**`.
   Without separators, `*` and `**` are the same.
5. `**` does not consume the separator next to it. With the separator `/`,
   `a/**/b` matches `a//b` and `a/x/b`, and not `a/b`. `**/x` does not match
   `x`. The pattern `{**/,}x` matches both (D16).
6. A pattern can hold U+FFFD. A byte that is not valid UTF-8 is an error (D8).

### Globstar

With `WithGlobstar`, a `**` that is a whole segment, and that a separator
follows, matches no segment too (D16):

1. The `**` is a whole segment when it starts the pattern, when it follows a
   separator, or when it follows a group whose patterns each end with a
   separator. `**/x` matches `x`, `a/**/b` matches `a/b`, and `**/**/x`
   matches `x`. An escaped separator counts, so `a\/**\/b` matches `a/b`.
2. `a/**` does not match `a`, as gitignore has it. `a/**/` matches `a/`.
3. A `**` that is not a whole segment, as in `a**/b`, keeps its meaning.

doublestar and minimatch also match `a` with `a/**`. git, doublestar and
`globset` read a `**` that is not a whole segment as `*`.

### Bracket expressions

A bracket expression is `[`, an optional `!` or `^`, one or more members, and
`]` (D9).

1. `!` or `^` right after `[` negates the class. In another place, each is a
   character, so `[a^]` and `[!!]` hold the character.
2. `]` right after `[`, `[!` or `[^` is a member: `[]]`, `[]a]`, `[!]a]`. So
   `[]` and `[!]` are not closed, and are an error.
3. Members mix characters and ranges: `[a-cx-z_]`, `[125-79]`.
4. `-` is a character when it is first, last, or right after a range:
   `[a-c-e]` holds `a` to `c`, `-` and `e`. `[--0]` is the range from `-` to
   `0`.
5. `\` escapes the next rune: `[\]]`, `[\-]`, `[\\]`, `[\a-\c]`.
6. A range compares code points. A range whose low end is after its high end,
   such as `[z-a]`, is an error.
7. A class matches one rune, and it can match a separator, as in upstream, Go
   and doublestar. The vectors with the feature `class-separator` hold this.
8. `[:name:]` is a named class. The names are `alnum`, `alpha`, `ascii`,
   `blank`, `cntrl`, `digit`, `graph`, `lower`, `print`, `punct`, `space`,
   `upper`, `word` and `xdigit`. `digit` and `xdigit` are ASCII. The others
   follow Unicode. An unknown name, or a name in upper case, is an error. A
   `[:` with no `:]` after it is the character `[`.
9. `[=c=]` and `[.c.]` with one rune are that rune. A longer name is an error.
10. A class that is not closed is an error.

### Escapes

1. `\c` outside a class matches `c`, for any rune. `\/` is a separator for
   the globstar.
2. A `\` at the end of the pattern, or of a class, is an error.

### Groups of alternatives

1. `{p1,p2}` matches when one of its patterns matches. A pattern in a group
   can hold any syntax, separators and groups too.
2. `{a,}`, `{,a}` and `{,}` hold the empty pattern. `{}` matches the empty
   string, and `{a}` is a group of one pattern. bash, minimatch and
   editorconfig read both as text.
3. `,` and `}` outside a group are characters.
4. A group that is not closed is an error. `,` and `}` in a class are members
   of the class.

### Brace sequences

A sequence is the whole body of a group, in the form `lo..hi` or
`lo..hi..step` (D11).

1. `{M..N}` with integers matches each integer from M to N, in either
   direction. A sign is allowed. `{1..10}` does not match `01`.
2. If an end has a leading zero, each item has the width of the wider end:
   `{01..10}` matches `01` to `10`, and `{1..010}` matches `001` to `010`.
3. The step is the absolute value of S, and 0 is 1. The sequence counts from M
   toward N: `{10..1..3}` matches 10, 7, 4 and 1.
4. `{a..e}` matches single runes by code point, in either direction, with an
   optional step. Each rune is a character, so `{A..c}` matches `[` and `\`.
5. Any other body is a group: `{1..a}`, `{aa..cc}`, `{1...3}` and `{1..}`
   match their text. `{1..2,x}` is a group of `1..2` and `x`. A sequence can
   be inside a group, as in `{x,{1..3}}`. `[{1..3}]` is a class, and `\{1..3}`
   is text.
6. A sequence of more than 10,000 items is an error.

### Extended groups

`?(l)`, `*(l)`, `+(l)` and `@(l)`, and `!(l)`, where `l` is patterns separated
by `|`, match as in bash with `extglob` (D10).

1. `?` matches zero or one pattern of the group, `*` zero or more, `+` one or
   more, and `@` exactly one. A pattern can be empty: `@()` matches the empty
   string only.
2. `!(l)` matches a run, the empty run too, that no pattern of the group
   matches and that holds no separator. `!(a)*` matches `a`, because `!(a)`
   matches the empty run. `!(!(a))` is `@(a)`.
3. `a*!(z)` matches `az`, as in picomatch. bash and glibc do not match it,
   because their `*` never leaves the empty run to the `!(`.
4. The other groups cross a separator only when a pattern of the group holds
   one: `?(x/)y` matches `x/y`.
5. `(`, `|` and `)` outside a group are characters. `\@(a)` is text. `**(` is
   a star and the group `*(`.
6. A group that is not closed, such as `@(a|b`, is an error.

### Case

With `WithFold`, a pattern matches as `strings.EqualFold` compares (D17). `K`
matches `k` and the Kelvin sign U+212A. `straße` does not match `STRASSE`,
because `ß` folds to `ẞ` and not to `SS`. A class matches a rune when it holds
the rune or another case of it: `[a-z]` matches `A`, and `[[:upper:]]`
matches `a`. git agrees. bash and glibc do not fold a named class or a range
whose ends differ in case.

## Where implementations disagree

| Case | The fork | Others |
| --- | --- | --- |
| `[]a]`, `[-x]`, `[x-]`, `[a-b-c]` | valid, as in bash | Go `path.Match`: an error |
| `[^...]` | a negation | Python, editorconfig and the fork before D9: the character `^` |
| an unclosed `[`, `{` or `(`, or a trailing `\` | an error, as in upstream, Go and `globset` | bash, minimatch, Python: text. git, glibc: no match |
| `[z-a]` | an error | bash, glibc, git: no match |
| a class and a separator, `a[/]b` and `a[!x]b` against `a/b` | a match, as in upstream, Go and doublestar | `FNM_PATHNAME`, git, `globset`, picomatch: no match |
| `{a}` and `{}` | a group, and the empty string | bash, minimatch, editorconfig: text |
| `foo{,.txt}` against `foo` | a match | `globset`: no match without `empty_alternates` |
| `a,b}` | text | `globset`: an error |
| `**/x` against `x`, and `a/**/b` against `a/b` | no match without `WithGlobstar` | doublestar, `globset`, minimatch, picomatch, git: a match |
| `a/**` against `a`, with the globstar | no match, as in upstream and gitignore | doublestar, minimatch: a match |
| `foo**bar` against `foo/baz/bar` | a match | doublestar, git, `globset`: no match |
| `a*!(z)` against `az` | a match | bash, glibc: no match |
| `!(a)*` against `a`, and `*(0|1)` against the empty string | a match, as in bash | picomatch: no match |
| `c!(z)` against `c/z`, with `/` | no match, as in minimatch | glibc: a match |
| `[[:word:]]` | letters, digits and `_` | glibc: an unknown class |
| a named class or a range with `WithFold` | fold, then compare | bash, glibc: no fold |
| `[[.hyphen.]]` | an error | bash: matches `-` |

## What the fork leaves out

`docs/BACKLOG.md` holds each item, and the vectors of each are in
`testdata/vectors.tsv`, which `TestVectors` skips:

1. A leading dot that only a literal `.` matches, as `FNM_PERIOD` and the
   shells do, at the start of the input and after each separator.
2. A match of a leading part of the input that ends before a separator, as
   `FNM_LEADING_DIR` does.
3. A `\` that is a character, as `FNM_NOESCAPE` has it.
4. A `!` at the start of a pattern that negates the whole pattern, as
   minimatch has it.
5. A pattern with no separator that matches the base name of the input, as
   the option `matchBase` of minimatch does.
6. The dialects of gitignore and editorconfig, and `fnmatch` of Python.
