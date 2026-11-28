# Decisions

Every decision of this project is a file in this folder, named by its number
and its title (D3). This table is the index. Find the number here, then open
the file.

Each file opens with its status. `Decided` means that the maintainers chose
it. `Proposed` means that it waits for the maintainers. A decision that
changes an earlier one says so in its status, as `Amends D3`, and the earlier
one says it back, as `Amended by D9`. Read the status before the decision.

A bare number, such as D3, names a decision in this folder. A decision of
another repository names that repository, such as dbmeta D110 or magic D23.

A new decision gets the next number and a file of its own. Add its row here.
`TestTheDecisionIndexIsComplete` prints the row when it is missing. An open
question is not a decision. It goes at the end of [`../PLAN.md`](../PLAN.md)
until the maintainers answer it.

D1 to D6 record the fork and the layout of the repository. D7 to D19 came
with the work that followed upstream and surveyed the syntax, and D20 came
with the work on the time and the allocations of a match. They were proposed
on 2026-11-26, and the maintainers decided them on 2026-11-28.

| # | Decision | Status |
| --- | --- | --- |
| D1 | [glob is a fork of gobwas/glob](D001-glob-is-a-fork-of-gobwas-glob.md) | Decided |
| D2 | [glob is set up for coding agents as every xo repository is](D002-glob-is-set-up-for-coding-agents-as-every-xo-repository-is.md) | Decided |
| D3 | [Each decision is a file of its own](D003-each-decision-is-a-file-of-its-own.md) | Decided |
| D4 | [The documents name no maintainer](D004-the-documents-name-no-maintainer.md) | Decided |
| D5 | [The module depends on the standard library alone](D005-the-module-depends-on-the-standard-library-alone.md) | Decided |
| D6 | [The upstream checkout is in upstream/, with a base commit and a ledger](D006-the-upstream-checkout-is-in-upstream-with-a-base-commit-and-a-ledger.md) | Decided |
| D7 | [The fork ports the behavior of upstream, not its engine](D007-the-fork-ports-the-behavior-of-upstream-not-its-engine.md) | Decided |
| D8 | [A malformed pattern is an error](D008-a-malformed-pattern-is-an-error.md) | Decided |
| D9 | [A bracket expression follows POSIX](D009-a-bracket-expression-follows-posix.md) | Decided |
| D10 | [Extended groups follow bash extglob, and are on by default](D010-extended-groups-follow-bash-extglob-and-are-on-by-default.md) | Decided |
| D11 | [A brace sequence follows bash](D011-a-brace-sequence-follows-bash.md) | Decided |
| D12 | [The backtracking matcher matches what the optimizer cannot, and is the reference of the tests](D012-the-backtracking-matcher-matches-what-the-optimizer-cannot-and-is-the-reference-of-the-tests.md) | Decided. Amended by D20 |
| D13 | [Options are functions, and Compile keeps its signature](D013-options-are-functions-and-compile-keeps-its-signature.md) | Decided |
| D14 | [Path and Host are globs with their own options](D014-path-and-host-are-globs-with-their-own-options.md) | Decided |
| D15 | [Quote escapes every special character, the comma too](D015-quote-escapes-every-special-character-the-comma-too.md) | Decided |
| D16 | [A `**` is a star that crosses separators, and globstar is an option](D016-a-is-a-star-that-crosses-separators-and-globstar-is-an-option.md) | Decided |
| D17 | [Case folding maps each rune to one rune of its case orbit](D017-case-folding-maps-each-rune-to-one-rune-of-its-case-orbit.md) | Decided |
| D18 | [The search of the optimizer has a budget](D018-the-search-of-the-optimizer-has-a-budget.md) | Decided |
| D19 | [The tests read vectors from the survey of glob syntax](D019-the-tests-read-vectors-from-the-survey-of-glob-syntax.md) | Decided |
| D20 | [A match takes linear time, and the NFA matcher replaces the backtracking matcher](D020-a-match-takes-linear-time-and-the-nfa-matcher-replaces-the-backtracking-matcher.md) | Decided. Amends D12 |
