# D2. glob is set up for coding agents as every xo repository is

Status: Decided.

The maintainers decided that every repository in the `xo` namespace is set up
for coding agents in the same way, and dbmeta records it as dbmeta D110. The
maintainers asked on 2026-11-26 for the same layout here, with dbmeta as a
template and not a copy. The repository now has these files:

- `AGENTS.md` holds the rules for a coding agent, and opens with the three
  standing rules of dbmeta D110.
- `CLAUDE.md` holds one line, `@AGENTS.md`, so that Claude Code reads the same
  rules. It is a file and not a symbolic link.
- `CONTRIBUTING.md` holds the same rules for a person, and the commands that
  install the agent skills.
- `skills-lock.json` names the two skills, `simple-english` and `go-pedantry`.
  Each is an ordinary folder in `.agents/skills/` and in `.claude/skills/`,
  installed with `--copy`.
- `.gitignore` ignores `.claude/settings.local.json`, the Claude Code
  permissions of one person.
- `.gitattributes` holds `* text=auto eol=lf`.
- `docs/` holds every other document: the plan, the backlog, the rules for
  following upstream, the survey of the syntax, and the decisions.

The tests in `skills_test.go` and `docs_test.go` hold the setup. They use only
the standard library.
