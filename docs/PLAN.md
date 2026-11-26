# Plan

This document holds the state of the work and the open questions. The
decisions are in `decisions/`, and the work that is known and not done is in
`BACKLOG.md`.

## State on 2026-11-26

The work of 2026-11-26 is staged and not committed. It did four things:

1. It set the repository up for coding agents in the layout of the `xo`
   repositories, with dbmeta as the template (D2 to D5).
2. It read upstream, which became active again in 2026-08 with a new engine
   and the releases `v1.0.0` and `v1.1.0`. It made a ledger of the commits of
   upstream (D6), and a table of the issues and pull requests of upstream in
   `UPSTREAM.md`. It ported the behavior of each fix that applied (D7, D8).
3. It surveyed the glob syntax of eleven other implementations
   (`SYNTAX.md`), and added what the survey found missing: full bracket
   expressions and named classes (D9), extended groups (D10), brace sequences
   (D11), case folding (D17), the globstar option (D16), and the types `Path`
   and `Host` from `notes.txt` (D14).
4. It added the backtracking matcher (D12), and a fuzz test that compares it
   with the optimized matchers. The survey and the fuzz test found seven
   faults of the optimizer, and each is fixed (D12).

`TestVectors` runs 2,029 of the 2,336 vectors of the survey, and each one
passes on both matchers. It skips the vectors of the options and the dialects
that the fork does not have.

## Open questions

The numbers do not repeat. The next question is question 6.

1. D7 to D19 are proposed. Do the maintainers decide them as they are?
2. Extended groups are on by default (D10). A pattern that held `@(` or `+(`
   as text now holds a group. Is that change acceptable for a minor version,
   or do the maintainers want an option that turns the groups on?
3. `[^...]` is a negation now (D9), and before it matched `^`. Is that change
   acceptable for a minor version?
4. The next tag. The API grew, and some patterns match differently (D8, D9,
   D10, D16). Is it `v0.2.0`?
5. `notes.txt` in the root holds the two ideas that D10 and D14 answer. git
   does not track it. Can a maintainer delete it?
