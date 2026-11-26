# D1. glob is a fork of gobwas/glob

Status: Decided.

`github.com/kenshaw/glob` is a fork of `github.com/gobwas/glob`. The fork
keeps the history of upstream to commit `e7a84e9`, the tip of master in
2018, and one more commit, `23a35a4`, which upstream does not hold on any
branch today. Upstream holds a commit with the same subject and date,
`e4652bc`, on the branch `feature/v0.3`, with different content.

From 2025-05-04, the fork moved the packages of upstream into the package
`syntax`, changed the API to match the other packages of the maintainers,
and added `encoding.TextMarshaler` and `encoding.TextUnmarshaler`. The fork
has the tags `v0.1.0` and `v0.1.1`.

The `LICENSE` of upstream, the MIT license, stays, and it names its holder.
