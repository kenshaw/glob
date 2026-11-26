# Contributing

This document is for a person who changes `glob`. `AGENTS.md` holds the same
rules for a coding agent, in more detail. `docs/decisions/` holds every
decision and the reason for it, and `docs/PLAN.md` holds the state of the
work and the open questions.

## Build and test

The package is pure Go. Run these commands in the repository root:

```bash
gofmt -l . && go vet ./... && go build ./... && go test -race -count=2 ./...
```

```bash
golangci-lint run ./...
```

`-race` needs cgo. Without a C compiler, run the tests with
`CGO_ENABLED=0 go test ./...`. `gofmt -l .` prints nothing when every file is
formatted. CI runs the same commands.

After a change of a matcher, of the optimizer or of the parser, run the fuzz
test for some minutes. It compares the optimized matchers with the
backtracking matcher (D12):

```bash
go test -run=NONE -fuzz=FuzzMatchersAgree -fuzztime=5m .
```

## Upstream

The fork follows `github.com/gobwas/glob` by behavior (D7).
`docs/UPSTREAM.md` holds the rules, and the ledger in `docs/upstream/`
records each commit of upstream. The checkout of upstream is in
`upstream/glob/`, which git ignores. Make it with:

```bash
git clone https://github.com/gobwas/glob.git upstream/glob
```

Without the checkout, `TestTheLedgerIsComplete` checks only the format of
the ledger.

## Documents

The root holds four documents: `README.md`, `AGENTS.md`, `CLAUDE.md` and this
one. Every other document goes in `docs/`. A decision is a file of its own in
`docs/decisions/`, with a row in the index `docs/decisions/README.md` (D3).
The tests in `docs_test.go` fail if the index and the files disagree. The
documents name no maintainer (D4).

## Agent skills

The repository carries two agent skills. A skill is a set of instructions
that a coding agent loads for a task. `simple-english` sets how text is
written, and `go-pedantry` sets how Go is written.

`skills-lock.json` names the source of each skill. The `skills` command from
npm writes that file, and version 1.7.0 is the one used here. It writes each
skill into two folders. Codex and the other agents read
`.agents/skills/<name>`, and Claude Code reads `.claude/skills/<name>`.

To install the skills again or to update them, run these commands in the
repository root:

```bash
npx skills@1.7.0 add AminBlg/SimpleEnglish --skill simple-english --agent codex claude-code --copy -y
```

```bash
npx skills@1.7.0 add oborchers/fractional-cto --skill go-pedantry --agent codex claude-code --copy -y
```

Keep `--copy`. Without it, the command writes `.claude/skills/<name>` as a
symbolic link. A Windows checkout writes a symbolic link as a small text
file, and Claude Code then loads no skill and reports nothing.
`TestSkillsAreCopies` fails on a link, on a missing copy, on two copies that
differ and on a skill that `skills-lock.json` does not name. See D2.

## Release

Tag only when a maintainer says so, and only after CI passes on the commit.
The module has a plain semver tag, such as `v0.2.0`. Never move a tag that is
published, because the module proxy caches it.
