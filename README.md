# glob

Package `glob` provides [standard wildcard][wildcards] matching for Go.

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

## Syntax

Syntax is inspired by [standard wildcards][wildcards], except that `**` is aka
super-asterisk, that do not sensitive for separators.

## Install

```sh
$ go get github.com/kenshaw/glob
```

## Example

```go

```

## Performance

This library is created for compile-once patterns. This means, that compilation
could take time, but strings matching is done faster, than in case when always
parsing template.

If you will not use compiled `glob.Glob` object, and do `g :=
glob.MustCompile(pattern); g.Match(...)` every time, then your code will be
much more slower.

Run `go test -bench=.` from source root to see the benchmarks:

| Pattern                                             | Fixture                       | Match   | Speed (ns/op) |
| --------------------------------------------------- | ----------------------------- | ------- | ------------- |
| `[a-z][!a-x]*cat*[h][!b]*eyes*`                     | `my cat has very bright eyes` | `true`  | 432           |
| `[a-z][!a-x]*cat*[h][!b]*eyes*`                     | `my dog has very bright eyes` | `false` | 199           |
| `https://*.google.*`                                | `https://account.google.com`  | `true`  | 96            |
| `https://*.google.*`                                | `https://google.com`          | `false` | 66            |
| `{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}`  | `http://yahoo.com`            | `true`  | 163           |
| `{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}`  | `http://google.com`           | `false` | 197           |
| `{https://*example.com,http://exclude.example.com}` | `https://safe.example.com`    | `true`  | 22            |
| `{https://*example.com,http://exclude.example.com}` | `http://safe.example.com`     | `false` | 24            |
| `abc*`                                              | `abcdef`                      | `true`  | 8.15          |
| `abc*`                                              | `af`                          | `false` | 5.68          |
| `*def`                                              | `abcdef`                      | `true`  | 8.84          |
| `*def`                                              | `af`                          | `false` | 5.74          |
| `ab*ef`                                             | `abcdef`                      | `true`  | 15.2          |
| `ab*ef`                                             | `af`                          | `false` | 10.4          |

The same things with `regexp` package:

| Pattern                                                               | Fixture                       | Match   | Speed (ns/op) |
| --------------------------------------------------------------------- | ----------------------------- | ------- | ------------- |
| `^[a-z][^a-x].*cat.*[h][^b].*eyes.*$`                                 | `my cat has very bright eyes` | `true`  | 2553          |
| `^[a-z][^a-x].*cat.*[h][^b].*eyes.*$`                                 | `my dog has very bright eyes` | `false` | 1383          |
| `^https:\/\/.*\.google\..*$`                                          | `https://account.google.com`  | `true`  | 1205          |
| `^https:\/\/.*\.google\..*$`                                          | `https://google.com`          | `false` | 767           |
| `^(https:\/\/.*\.google\..*\|.*yandex\..*\|.*yahoo\..*\|.*mail\.ru)$` | `http://yahoo.com`            | `true`  | 1435          |
| `^(https:\/\/.*\.google\..*\|.*yandex\..*\|.*yahoo\..*\|.*mail\.ru)$` | `http://google.com`           | `false` | 1674          |
| `^(https:\/\/.*example\.com\|http://exclude.example.com)$`            | `https://safe.example.com`    | `true`  | 1039          |
| `^(https:\/\/.*example\.com\|http://exclude.example.com)$`            | `http://safe.example.com`     | `false` | 272           |
| `^abc.*$`                                                             | `abcdef`                      | `true`  | 237           |
| `^abc.*$`                                                             | `af`                          | `false` | 100           |
| `^.*def$`                                                             | `abcdef`                      | `true`  | 464           |
| `^.*def$`                                                             | `af`                          | `false` | 265           |
| `^ab.*ef$`                                                            | `abcdef`                      | `true`  | 375           |
| `^ab.*ef$`                                                            | `af`                          | `false` | 145           |

Syntax is inspired by [standard wildcards][wildcards], except that `**` is aka
super-asterisk, that do not sensitive for separators.

[wildcards]: http://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm
