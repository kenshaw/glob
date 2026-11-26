# glob

`glob` is a Go package that matches strings against glob patterns, such as
`*.go`, `a/**/b` or `{https,http}://*.example.com`. It is a fork of
`github.com/gobwas/glob` (D1). It compiles a pattern once to a tree of
matchers, and the matchers answer without a regular expression.

The fork follows upstream by behavior, not by code (D7). `docs/UPSTREAM.md`
holds the rules and the answer of the fork to each issue of upstream. The
syntax follows the shells, POSIX and upstream, and `docs/SYNTAX.md` holds
the survey that chose it.

## Standing rules

These rules hold in every `xo` repository, for every coding agent (dbmeta
D110, which D2 adopts).

1. Stage changes for review. Commit and push only when a maintainer says so.
2. Load the `simple-english` skill before you write any text that a person
   reads: project documentation, a code comment, an error message or a commit
   message.
3. Load the `go-pedantry` skill before you write or review Go code. A rule in
   this file wins where the two conflict.

`CLAUDE.md` holds one line that imports this file, so that Claude Code and
every other agent read the same rules. Edit this file, not `CLAUDE.md`.

## Which document to read

`docs/PLAN.md` holds the state of the work and the open questions.
`docs/decisions/` holds every decision, one file each (D3), and
`docs/decisions/README.md` is the index. Read the status of a decision before
the decision, because some are only proposed. Do not decide an open question
on your own. The open questions are at the end of `docs/PLAN.md`. Ask a
maintainer.

| If you are | Read |
| --- | --- |
| asking why something is the way it is | the index in `docs/decisions/README.md` |
| changing the syntax of a pattern | `docs/SYNTAX.md`, then D8 to D11 and D16 |
| changing the API in `glob.go` | D13 and D14, then Go conventions below |
| changing a matcher or the optimizer | D12 and D18. Run `FuzzMatchersAgree` after the change |
| following upstream, or adding to the ledger | `docs/UPSTREAM.md`, D6 and D7 |
| answering an issue of upstream | the tables in `docs/UPSTREAM.md`, then `TestUpstreamIssues` |
| adding a test vector | D19, then `testdata/vectors.tsv` |
| changing CI | `.github/workflows/test.yml` |
| looking for work that is known and not done | `docs/BACKLOG.md` |
| writing a document, a code comment, an error message or a commit message | the `simple-english` skill. Load it first |
| writing or reviewing Go code | the `go-pedantry` skill, then Go conventions below |
| adding or updating an agent skill | `CONTRIBUTING.md`, under Agent skills, and D2 |

`README.md` is for a person who uses the package. `CONTRIBUTING.md` is for a
person who changes it.

## How to name a decision

A bare number, such as D3, names a decision of this repository, in
`docs/decisions/`. A decision of another repository is named with that
repository, such as dbmeta D110 or magic D23.
`TestEveryDecisionReferenceExists` fails on a bare number that
`docs/decisions/` does not hold.

A new decision gets the next number and a file of its own,
`docs/decisions/D<nnn>-<title>.md`, and a row in `docs/decisions/README.md`.
The file opens with `# D<n>. <title>`, a blank line, and `Status: <status>.`.
The status is `Decided` or `Proposed`. A decision that an agent writes is
`Proposed` until a maintainer decides it. If it changes an earlier decision,
its status also says `Amends Dn`, and the status of the earlier one says
`Amended by D<n>`. `TestTheDecisionIndexIsComplete` and
`TestAnAmendmentPointsBothWays` hold this.

An open question has a number, and it is at the end of `docs/PLAN.md` until a
maintainer answers it. The numbers do not repeat.

The documents name no maintainer (D4). Write "the maintainers" for a choice
and "a maintainer" for a request. Write "upstream" or "the reporter" for a
person of upstream, and name an issue by its number.

## Layout

| Path | What it holds |
| --- | --- |
| `glob.go` | the API: `Compile`, `CompileWith`, `Must`, `New`, `Quote`, the options, and the types `Glob`, `Path` and `Host` (D13, D14) |
| `syntax/` | the lexer, the parser, the tree, the optimizer and the matchers |
| `syntax/lexer.go`, `syntax/parser.go` | the pattern to tokens, and the tokens to a tree |
| `syntax/node.go`, `syntax/syntax.go` | the tree, and its minimizer |
| `syntax/class.go` | bracket expressions, named classes, brace sequences, case folding and globstar (D9, D11, D16, D17) |
| `syntax/match.go`, `syntax/matcher.go` | the optimizer and the optimized matchers |
| `syntax/backtrack.go` | the backtracking matcher, for extended groups and as the reference of the tests (D10, D12) |
| `glob_test.go`, `example_test.go`, `bench_test.go` | the tests, the examples and the benchmarks of the API |
| `diff_test.go`, `vectors_test.go` | `FuzzMatchersAgree`, `FuzzQuote` and `TestVectors` (D12, D19) |
| `skills_test.go`, `docs_test.go`, `ledger_test.go`, `gomod_test.go` | the tests of the agent setup, the documents, the ledger and `go.mod`, in package `glob_test` |
| `testdata/vectors.tsv` | the test vectors of the survey (D19) |
| `testdata/fuzz/` | the inputs that the fuzz tests failed on once |
| `cmd/globtest`, `cmd/globdraw` | programs that test a pattern and draw its matchers |
| `bench.sh` | compares the benchmarks of two revisions with benchstat |
| `docs/PLAN.md` | the state of the work and the open questions |
| `docs/SYNTAX.md` | the survey of glob syntax and the choices of the fork |
| `docs/UPSTREAM.md` | the rules for following upstream, and the issues of upstream |
| `docs/upstream/` | the list of upstreams and the ledger of upstream commits (D6) |
| `docs/BACKLOG.md` | the work that is known and not done |
| `docs/decisions/` | every decision, one file each, and the index in `README.md` (D3) |
| `upstream/glob/` | the checkout of upstream. git ignores it. Do not edit it (D6) |

## Commands

The package is pure Go. Run these commands in the repository root:

```bash
gofmt -l . && go vet ./... && go build ./... && go test -race -count=2 ./...
```

```bash
golangci-lint run ./...
```

`gofmt -l .` prints nothing when every file is formatted. After a change of a
matcher, of the optimizer or of the parser, run the fuzz test for some
minutes. It must find nothing:

```bash
go test -run=NONE -fuzz=FuzzMatchersAgree -fuzztime=5m .
```

If it finds an input, it writes the input to `testdata/fuzz/`. Fix the fault,
and keep the file, so that `go test` runs the input every time.

## Go conventions

- Wrap every error with `%w`. Write an error message in lower case, and do not
  start it with "failed to" or "error".
- A matcher keeps no state between calls, and keeps no input after a call.
  Every matcher is safe for concurrent use. `MatchBytes` depends on this.
- The optimized matchers and the backtracking matcher must agree on every
  input (D12). A new optimization needs a fuzz run.
- A test that names an issue of upstream names it by its number, as in
  "issue 51 of upstream".
- `go.mod` has no `require` line (D5).
- Name a receiver with one or two lower case letters, the same on every
  method of a type.

The code of the package `syntax` came from upstream, and some of it differs
from `go-pedantry`: exported names without a comment, and names with an
underscore in the tests. `docs/BACKLOG.md` holds the item.

## Linting

`.golangci.yml` holds the rules. The posture is `default: all` with a disable
list. A linter that makes idiomatic Go worse is disabled, with the reason
written beside it. Only a real defect gets a code change. A change made to
quiet a linter, rather than to fix something, is itself a defect.
