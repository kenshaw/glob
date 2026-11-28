# Plan

This document holds the state of the work and the open questions. The
decisions are in `decisions/`, and the work that is known and not done is in
`BACKLOG.md`.

## State on 2026-11-26

The work of 2026-11-26 is in `v0.2.0`. It did four things:

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

Later on 2026-11-26, the work on the time and the allocations of a match did
four things (D20):

1. It found that a match could take time of the order of the length of the
   input to the power of the number of stars, and made each match linear. The
   NFA matcher replaces the backtracking matcher, and the compiler gives it
   each tree of optimized matchers that `syntax.Linear` does not know to be
   linear.
2. It removed the allocations of a match. A match with fold does not copy an
   input that is lower case already (D17), and `MatchBytes` of `Path` and
   `Host` does not copy its input. A glob without a pattern matches nothing,
   and does not panic.
3. It added `FuzzCompile`, and sets of separators to `FuzzMatchersAgree`.
   The fuzz tests and a reading of the code found three faults on input that
   is not valid UTF-8, two faults with separators, and a debug build that did
   not compile. Each is fixed, and `UPSTREAM.md` lists them.
4. It made the compiler faster: `Compile` of the benchmark pattern takes 5.4
   microseconds and 183 allocations, in place of 6.7 microseconds and 293
   allocations.

A maintainer asked for the minor release `v0.2.0` on 2026-11-26, which
answers question 4. It holds both pieces of work.

On 2026-11-28, the maintainers decided D7 to D20 as they were proposed. That
answers questions 1, 2 and 3: the extended groups stay on by default, and
`[^...]` stays a negation, in a minor version.

## Open questions

The numbers do not repeat. The next question is question 7.

5. `notes.txt` in the root holds the two ideas that D10 and D14 answer. git
   does not track it. Can a maintainer delete it?
6. A pattern whose tree of optimized matchers is not linear gets the NFA
   matcher (D20). It is safe on any input, but it can take about twice the
   time on a short input. The benchmark pattern
   `[a-z][!a-x]*cat*[h][!b]*eyes*` takes 180 ns in place of 90 ns. Is that
   cost acceptable, or do the maintainers want an option that keeps the
   optimized matchers for trusted input?
