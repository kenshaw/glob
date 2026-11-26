# Backlog

This document holds the work that is known and not done. Each item names
where it came from.

When you finish an item, remove it. If an item needs a decision, it becomes
an open question at the end of `PLAN.md`, and the answer becomes a file in
`decisions/`.

## Syntax

1. An option for a leading dot, as `FNM_PERIOD` and the shells have it. The
   vectors of `opt-period` hold the meaning. Source: `SYNTAX.md`.
2. An option for `FNM_LEADING_DIR`, an option for `FNM_NOESCAPE`, the
   negation of a whole pattern, and the match of a base name. The vectors of
   `opt-leadingdir`, `opt-noescape`, `opt-negate` and `opt-matchbase` hold
   the meaning. Source: `SYNTAX.md`.
3. Types for the dialects of gitignore and editorconfig, as `Path` and
   `Host` are for paths and host names. The vectors of `dialect-gitignore`
   and `dialect-editorconfig` hold the meaning. Source: `SYNTAX.md`.
4. An error type with the offset in the pattern, as `*SyntaxError` of
   upstream has. Today an error is a string. Source: `c0ec127` of upstream.

## Matchers

5. The optimizer does not hold extended groups, so a pattern with a group
   gets the backtracking matcher (D10, D12). It takes cubic time in the
   length of the input at worst. A group with one pattern and no repetition,
   such as `@(a|b)`, can become a group of alternatives and use the
   optimizer.
6. The optimizer tries every way to glue neighbors, within a budget (D18). A
   search that is linear, glueing each run of neighbors that can glue, can
   give the same result for most patterns.
7. A matcher of a brace sequence for a long range of numbers. Today the parser
   expands it to alternatives, up to 10,000 (D11).

## Code

8. Much of the package `syntax` came from upstream, and differs from
   `go-pedantry`: exported names have no comment, the tests use names with an
   underscore, and `Node.String` uses `bytes.Buffer`. `golangci-lint` reports
   these, and `.golangci.yml` disables the linters with that reason. Source:
   D2.
9. The package `syntax` exports its tokens, nodes and matchers. A program
   needs only the package `glob`. Moving `syntax` to `internal/syntax` is a
   change of the API, and needs a decision.
