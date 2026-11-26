# glob

Package `glob` provides [glob wildcard][wildcards] matching for Go, and is a
fork of the [`github.com/gobwas/glob`][gobwas] package.

The fork has the syntax of the shells: bracket expressions with named classes
such as `[[:alpha:]]`, the extended groups of bash such as `@(a|b)` and
`!(a)`, and brace sequences such as `{1..10}`. It has options to match without
regard to case and to give `**` the meaning of the globstar of a shell, and
the types `Path` and `Host`. It depends on the standard library alone.

A match takes time linear in the length of the input, so a glob is safe to
match against input that you do not trust. A group of bash that repeats a
pattern with no most length, such as `+(*a)`, can take quadratic time. A match
of most patterns does not allocate.

[![Unit Tests][glob-ci-status]][glob-ci]
[![Go Reference][goref-glob-status]][goref-glob]
[![Releases][release-status]][Releases]
[![Discord Discussion][discord-status]][discord]

[glob-ci]: https://github.com/kenshaw/glob/actions/workflows/test.yml "Test CI"
[glob-ci-status]: https://github.com/kenshaw/glob/actions/workflows/test.yml/badge.svg "Test CI"
[goref-glob]: https://pkg.go.dev/github.com/kenshaw/glob "Go Reference"
[goref-glob-status]: https://pkg.go.dev/badge/github.com/kenshaw/glob.svg "Go Reference"
[release-status]: https://img.shields.io/github/v/release/kenshaw/glob?display_name=tag&sort=semver "Latest Release"
[discord]: https://discord.gg/WDWAgXwJqN "Discord Discussion"
[discord-status]: https://img.shields.io/discord/829150509658013727.svg?label=Discord&logo=Discord&colorB=7289da&style=flat-square "Discord Discussion"
[releases]: https://github.com/kenshaw/glob/releases "Releases"

## Install

```sh
$ go get github.com/kenshaw/glob
```

## Example

The package can be used like the following:

```go
import "github.com/kenshaw/glob"

g, err := glob.Compile(`test*`)
if err != nil {
    /* ... */
}
fmt.Println(g.Match("test_file.txt"))

// Output:
// true
```

See [the package examples][pkg-overview] for more examples on using `glob`.

Options go to `CompileWith`:

```go
g, err := glob.CompileWith(`*.GO`, glob.WithFold())
// g.Match("main.go") == true

p, err := glob.CompilePath(`src/**/*.go`)
// p.Match("src/main.go") == true

h, err := glob.CompileHost(`*.example.com`)
// h.Match("WWW.Example.COM.") == true
```

A struct field of type `glob.Glob`, `glob.Path` or `glob.Host` unmarshals
from a pattern in JSON, YAML or any other format of text.

## Syntax

Syntax is meant to be compatible with [standard wildcards][wildcards] (as used
in Bash or other shells). Additionally, `glob` supports the "super star" (`**`)
pattern, traversing path separators. A `**` does not consume the separator
next to it, so `a/**/b` with the separator `/` does not match `a/b`. The
option `WithGlobstar`, and the type `Path`, give `**/` the meaning of the
globstar of a shell, which matches no segment too.

| Pattern | Matches |
| :--- | :--- |
| `*` | any sequence of characters that are not separators |
| `**` | any sequence of characters, separators too |
| `?` | one character that is not a separator |
| `[abc]`, `[a-z0-9_]` | one character of the class |
| `[!a-z]`, `[^a-z]` | one character not in the class |
| `[[:alpha:]]` | one character of a named class: `alnum`, `alpha`, `ascii`, `blank`, `cntrl`, `digit`, `graph`, `lower`, `print`, `punct`, `space`, `upper`, `word`, `xdigit` |
| `{a,b,c}` | one of the patterns |
| `{1..10}`, `{01..10..2}`, `{a..e}` | one item of the sequence |
| `?(a\|b)`, `*(a\|b)`, `+(a\|b)`, `@(a\|b)` | zero or one, zero or more, one or more, or exactly one of the patterns |
| `!(a\|b)` | any sequence of characters that are not separators, that none of the patterns matches |
| `\c` | the character `c` |

A pattern that is not complete, such as `a{b` or `[a`, is an error.
[`docs/SYNTAX.md`](docs/SYNTAX.md) holds every rule and edge case, and how
other implementations differ.

Some examples:

| Pattern                                             | Test                          |   Match |
| :-------------------------------------------------- | :---------------------------- | ------: |
| `[a-z][!a-x]*cat*[h][!b]*eyes*`                     | `my cat has very bright eyes` |  `true` |
| `[a-z][!a-x]*cat*[h][!b]*eyes*`                     | `my dog has very bright eyes` | `false` |
| `https://*.google.*`                                | `https://account.google.com`  |  `true` |
| `https://*.google.*`                                | `https://google.com`          | `false` |
| `{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}`  | `http://yahoo.com`            |  `true` |
| `{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}`  | `http://google.com`           | `false` |
| `{https://*example.com,http://exclude.example.com}` | `https://safe.example.com`    |  `true` |
| `{https://*example.com,http://exclude.example.com}` | `http://safe.example.com`     | `false` |
| `abc*`                                              | `abcdef`                      |  `true` |
| `abc*`                                              | `af`                          | `false` |
| `*def`                                              | `abcdef`                      |  `true` |
| `*def`                                              | `af`                          | `false` |
| `ab*ef`                                             | `abcdef`                      |  `true` |
| `ab*ef`                                             | `af`                          | `false` |

## Contributing

[`CONTRIBUTING.md`](CONTRIBUTING.md) holds the commands that build and test
the package. [`docs/UPSTREAM.md`](docs/UPSTREAM.md) holds how the fork
follows upstream, and [`docs/decisions/`](docs/decisions/README.md) holds every
decision.

[gobwas]: https://github.com/gobwas/glob
[wildcards]: http://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm
[pkg-overview]: https://pkg.go.dev/github.com/kenshaw/glob#pkg-overview
