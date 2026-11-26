# Following upstream

This document holds the rules for following upstream, `github.com/gobwas/glob`,
and the answer of the fork to each issue and pull request of upstream. The
rules come from D6 and D7.

## The upstream

`docs/upstream/upstreams.tsv` lists the upstream, with these columns,
separated by a tab:

| Column | What it holds |
| --- | --- |
| `repository` | the short name, which is also the name of the ledger and of the checkout |
| `url` | the URL to clone |
| `checkout` | the folder of the checkout |
| `base` | the full hash of the base commit |
| `matched` | the full hash of the last commit of the ledger that the fork matches |

## The checkout

The checkout is in `upstream/glob/`, in the root of this repository, and git
ignores the folder `upstream/`. The checkout is a reference. Do not edit a file
in it, and do not commit it. If the checkout is missing, make it in the root
of the repository:

```bash
git clone https://github.com/gobwas/glob.git upstream/glob
```

Upstream is a Go module, so `go build ./...` in the root of this repository
does not reach the checkout: the checkout has its own `go.mod`.

## The history

The fork holds the history of upstream to the base commit, `e7a84e9`, the tip
of master in 2018. After it, the fork holds `23a35a4`, a commit that upstream
holds on no branch. Upstream has a commit with the same subject and date,
`e4652bc`, on the branch `feature/v0.3`, with different content. The fork then
rewrote the layout and the API (D1).

Upstream has these branches on 2026-11-26. The ledger follows master only.

| Branch | Tip | What it holds |
| --- | --- | --- |
| `master` | `ff0540d` | the new engine of `v1.0.0` and `v1.1.0`, which the ledger follows |
| `feature/v0.3` | `bcbbef0` | the refactoring that the fork started from, and 11 commits of work in progress from 2018 and 2019 |
| `feature/v1`, `feature/globstar`, `feature/tail-alts`, `feature/modernize-source` | | the work that became master in 2026 |
| `optimize`, `optimize-noacquire` | | experiments from 2016. The fork holds `optimize` |

## The ledger

`docs/upstream/glob.tsv` has one line for each commit of master after the base
commit, in the order of upstream, with these columns:

| Column | What it holds |
| --- | --- |
| `upstream` | the full hash of the upstream commit |
| `date` | the date of the commit, as `git log --format=%cs` prints it |
| `status` | `ported`, `not-applicable` or `pending` |
| `glob` | the short hash of the commit of the fork that ports it, or `staged` until that commit exists, or empty |
| `note` | the subject of the commit, and what the fork did with it |

`TestTheLedgerIsComplete` fails on a line that is out of order, on a gap, and
on a status that is not one of the three. It reads the checkout when it
exists, and skips that part when it does not.

To add the commits that arrived since the last time:

1. Fetch the checkout: `git -C upstream/glob fetch` and
   `git -C upstream/glob merge --ff-only origin/master`.
2. Add a line for each commit after the last line of the ledger.
3. Run `go test -run TestTheLedgerIsComplete .`.

## Porting a commit

The fork ports the behavior of upstream, not its code (D7).

1. Read the diff: `git -C upstream/glob show <hash>`.
2. Write a test that shows the behavior in the fork. If the test passes
   already, the commit is ported, with a note that says so.
3. Change the fork until the test passes. `FuzzMatchersAgree` must still pass.
4. Set the status to `ported`, and the column `glob` to `staged`. After the
   commit of the fork exists, put its short hash there.
5. Move `matched` in `upstreams.tsv` to this commit, if no earlier commit is
   still `pending`.

## Issues and pull requests

The table holds every issue and pull request of upstream that is open on
2026-11-26, and each closed one that names a fault. It names each one by its
number on upstream. The fork answers each one in a test, in a decision, or with
the reason that it does not apply. `TestUpstreamIssues` in `glob_test.go`
holds the cases of the issues.

### Faults that applied to the fork

| Upstream | Fault | Answer of the fork |
| --- | --- | --- |
| issues 51 and 61 | `a*ant` matched `ant`: a prefix and a suffix overlapped | fixed in `PrefixSuffixMatcher` (D12) |
| issues 58 and 76 | `a/**/b` matched `a/b` because of the same overlap | fixed. `WithGlobstar` gives the shell meaning (D16) |
| issues 56 and 59 | an unclosed `{` compiled, and `0{` matched `0` | an error (D8) |
| pull request 83 | `[\-a]` read the backslash as the low end of a range | fixed (D9) |
| `dab909e` | a trailing `\` was ignored, and U+FFFD in a class was an error | fixed (D8) |
| the fuzz corpus of upstream | a NUL byte ended the pattern | fixed (D8) |
| issue 78, pull requests 79 and 80 | `Quote` did not escape the comma | `Quote` escapes it (D15) |
| pull request 84 | a `}` outside a group is a character | kept as a character, as upstream keeps it (D8) |

### Faults that no issue of upstream names

The fork found these faults in the code that it holds from the base commit.
No issue of upstream names them. The engine of `v1.0.0` replaced the code, so
each one is a fault of upstream before `v1.0.0`.

| Fault | Answer of the fork |
| --- | --- |
| a tree of stars took time of the order of the length of the input to the power of the number of stars. `*a*a*a*a*b` took 40 seconds on 400 bytes | linear time (D20) |
| with more than one separator, a star found the first separator of the list, not the first separator of the input | fixed (D20) |
| the last separator that is not ASCII was found in the wrong part of the input, and could panic | fixed (D20) |
| a byte that is not valid UTF-8 had the width 3 in the segments of five matchers | fixed (D20) |
| the tail of an input counted a byte that is not valid UTF-8 wrongly, and took U+FFFD for such a byte | fixed (D20) |
| the debug build, with the tag `globdebug`, did not compile | fixed. CI vets it |

### Faults that the fork had fixed already

Issues 7, 9, 11, 15, 18, 23, 28, 33, 36, 38, 41, 43, 50, 54, 62 and 66, and
pull request 55, name faults that the fork does not have.
`TestUpstreamIssues` holds the cases of issues 33, 36, 41, 43, 50, 51, 54,
61, 62 and 66. Pull requests 68, 69, 71 and 72 change code that the fork
holds in a different form, where the fault does not happen.

### Features

| Upstream | Feature | Answer of the fork |
| --- | --- | --- |
| issue 3, pull request 77 | extended groups such as `@(a|b)` and `!(a)` | added (D10) |
| issue 47, pull request 67 | negation | `!(...)` (D10). The syntax `{!a,b}` of pull request 67 is not standard, and upstream refused it |
| issue 37 | matching without regard to case | `WithFold` (D13, D17) |
| issue 32, pull request 81 | `MatchBytes` | added (D13) |
| issues 46, 58 and 76, pull request 82 | globstar | `WithGlobstar` (D16) |
| `a82038c` | `Separators` | added (D13) |
| `fuzz_test.go` of upstream | a fuzz test against a reference | `FuzzMatchersAgree` and `FuzzCompile`, against the NFA matcher (D12, D20) |

### Not applicable

| Upstream | Reason |
| --- | --- |
| issue 42, pull requests 30 and 40 | the code that they change is gone from the fork |
| pull requests 26, 27, 48, 49, 70 and 75 | the build, CI and the documents of upstream. The fork has its own |
| pull requests 73, 74 and 82 | the new engine of upstream (D7). The fork ports the globstar of 82 as an option |
| pull request 31 | the branch `feature/v0.3`, which the fork replaced |
| pull request 10 | closed, and the fork has no `AnyOf` or `EveryOf` glob |
| pull request 64 | an example of `QuoteMeta`. The fork fixed the example of `Quote` |
| issues 1, 2, 6, 14, 19, 22, 24, 34, 44, 45, 57, 60 and 63 | questions or documents. The fork has `Quote`, `MarshalText` and errors from `Compile` |
| issues 20, 29, 35 and 46 | the intended meaning: `**` crosses a separator, `*` does not, and `\` escapes |
| issue 21 | benchmarks only |
