# D10. Extended groups follow bash extglob, and are on by default

Status: Proposed.

The pattern syntax holds the five extended groups of bash and ksh:
`?(a|b)`, `*(a|b)`, `+(a|b)`, `@(a|b)` and `!(a|b)`. Issue 3 of upstream asks
for them, and `notes.txt` of the maintainers names them.

1. The groups are on without an option, as minimatch has them. A pattern that
   held `@(` as text before now holds a group. `\(` or `Quote` keeps the
   text (D15).
2. A group holds patterns separated by `|`. A pattern in a group can hold any
   syntax, groups too.
3. `!(...)` matches a sequence that does not hold a separator, and that no
   pattern of the group matches. This is what the group matches in one path
   segment of bash. `*`, `?` and `+` groups cross a separator only when a
   pattern of the group holds one.
4. A pattern that holds a group gets the NFA matcher (D12, D20). The
   optimizer does not hold groups.
