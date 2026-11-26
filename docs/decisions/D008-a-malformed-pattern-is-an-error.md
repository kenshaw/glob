# D8. A malformed pattern is an error

Status: Proposed.

`Compile` returns an error for a pattern that is not complete:

- an unclosed `{`, `[` or extended group, such as `a{b` or `@(a`,
- a `\` at the end of the pattern,
- a byte that is not valid UTF-8,
- an unknown named class, such as `[[:alfa:]]`.

The fork accepted `a{b` before, and matched it as `ab`. It also ignored a `\`
at the end. Upstream made the same cases an error in its `v1.0.0`, issues 56
and 59 of upstream. A `}`, `)` or `|` outside a group is a character, as
upstream and the shells treat it.

The rune U+FFFD in a pattern is a character, and only a byte that is not
valid UTF-8 is an error. A NUL byte is a character, and no longer ends the
pattern.
