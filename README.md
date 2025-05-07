# glob

Package `glob` provides [glob wildcard][wildcards] matching for Go, and is a
fork of the [`github.com/gobwas/glob`][gobwas] package.

Eventual goal is a complete rewrite of the actual lexer/parser implementation.
Current progress is simply a complete package retooling, cleanup, and API
changes to match other packages.

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

## Syntax

Syntax is meant to be compatible with [standard wildcards][wildcards] (as used
in Bash or other shells). Additionally, `glob` supports the "super star" (`**`)
pattern, traversing path separators.

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

[gobwas]: https://github.com/gobwas/glob
[wildcards]: http://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm
[pkg-overview]: https://pkg.go.dev/github.com/kenshaw/glob#pkg-overview
