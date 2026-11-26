# D3. Each decision is a file of its own

Status: Decided.

Each decision is `docs/decisions/D<nnn>-<title>.md`, with the number in three
digits so that a listing sorts in order. The file opens with its number and
title, a blank line, and its status:

    # D7. The fork ports the behavior of upstream, not its engine

    Status: Proposed.

The status is `Decided` when the maintainers chose it, and `Proposed` when it
waits for the maintainers. A decision that changes an earlier one says
`Amends Dn`, and the earlier one says `Amended by Dn`.

[`README.md`](README.md) in this folder is the index.
`TestTheDecisionIndexIsComplete`, `TestAnAmendmentPointsBothWays` and
`TestEveryDecisionReferenceExists` read this folder.

This follows dbmeta D111 and magic D23.
