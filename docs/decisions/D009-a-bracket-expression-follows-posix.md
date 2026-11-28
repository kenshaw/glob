# D9. A bracket expression follows POSIX

Status: Decided.

A bracket expression, such as `[a-z0-9_]`, follows POSIX and the shells:

1. It holds any number of characters, ranges and named classes, in any order.
   Before, it held one range or one list of characters, and `[a-cx]` was an
   error.
2. `[!...]` and `[^...]` are both negations. Before, `[^a]` matched `^` or
   `a`.
3. A `]` that is the first character, after an optional `!` or `^`, is a
   character. A `-` that is the first or the last character is a character.
   A character escaped with `\` is a character, and `[\-a]` matches `-` or
   `a`, which pull request 83 of upstream asks for.
4. The named classes are the twelve of POSIX, `alnum`, `alpha`, `blank`,
   `cntrl`, `digit`, `graph`, `lower`, `print`, `punct`, `space`, `upper` and
   `xdigit`, and two of bash and of other tools, `ascii` and `word`. `digit`
   and `xdigit` match ASCII only. The others follow the Unicode categories,
   as a shell does in a UTF-8 locale.
5. An equivalence class, such as `[=a=]`, and a collating symbol, such as
   `[.-.]`, hold one character, and match that character.

A bracket expression with one range or one list of characters still parses
to the node `Range` or `List`, so the optimizer keeps its shortcuts. Every
other one parses to the node `Class`.
