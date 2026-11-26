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
   gets the NFA matcher (D10, D20). A group with one pattern and no
   repetition, such as `@(a|b)`, can become a group of alternatives and use
   the optimizer.
6. The optimizer tries every way to glue neighbors, within a budget (D18). A
   search that is linear, glueing each run of neighbors that can glue, can
   give the same result for most patterns.
7. A matcher of a brace sequence for a long range of numbers. Today the parser
   expands it to alternatives, up to 10,000 (D11).
8. A group `+(...)`, `*(...)` or `!(...)` whose patterns have no most width,
   such as `+(*a)`, takes quadratic time at worst on the NFA matcher (D20).
   A set of positions that holds only the positions that changed can make it
   linear.
9. `IndexedAnyOfMatcher.Index` searches each pattern of the group again from
   each offset, so a tree with such a value is not linear, and its pattern
   gets the NFA matcher (D20). `{a,b}*{c,d}` is one such pattern. An index
   that remembers the next occurrence of each pattern for the length of one
   call makes it linear.
10. The NFA matcher takes about twice the time of the optimized matchers on a
    pattern such as `[a-z][!a-x]*cat*[h][!b]*eyes*` (D20), whose tree is not
    linear. In that tree, the left side matches each longer string too, and
    the right side matches each string that ends with a string that it
    matches. So only the first occurrence of the value where the left side
    matches can count. A search for that occurrence in linear time makes the
    tree linear.
11. `releaseSegments` puts a slice into a `sync.Pool`. The pool boxes the
    slice header, so each put allocates 24 bytes. A pool of pointers to
    slices does not.

## Code

12. Much of the package `syntax` came from upstream, and differs from
    `go-pedantry`: exported names have no comment, the tests use names with
    an underscore, and `Node.String` uses `bytes.Buffer`. `golangci-lint`
    reports these, and `.golangci.yml` disables the linters with that reason.
    Source: D2.
13. The package `syntax` exports its tokens, nodes and matchers. A program
    needs only the package `glob`. Moving `syntax` to `internal/syntax` is a
    change of the API, and needs a decision.
