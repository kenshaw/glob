# D6. The upstream checkout is in upstream/, with a base commit and a ledger

Status: Decided.

1. The checkout of upstream is in `upstream/glob/`, in the root of this
   repository. `.gitignore` holds `/upstream/`. Nobody edits a file in it.
2. `docs/upstream/upstreams.tsv` lists the upstream: its URL, its checkout,
   the base commit, and the last commit that the fork matches.
3. The base commit is `e7a84e9525fe90abcda167b604e483cc959ad4aa`, the last
   commit that the fork and the master branch of upstream share.
4. The ledger is `docs/upstream/glob.tsv`. It has one line for each commit of
   the master branch of upstream after the base commit, in the order of
   upstream, with no gap, and a status for each: `ported`, `not-applicable`
   or `pending`. `TestTheLedgerIsComplete` makes sure of that when the checkout
   exists.
5. `docs/UPSTREAM.md` holds the rules, and the issues and pull requests of
   upstream with the answer of the fork to each.
6. The fork follows upstream when a maintainer asks. No scheduled task fetches
   upstream.

This follows magic D25 and magic D26.
