# D15. Quote escapes every special character, the comma too

Status: Proposed.

`Quote` escapes `*`, `?`, `\`, `[`, `]`, `{`, `}`, `,`, `(`, `)` and `|`. A glob
compiled from the result matches the input of `Quote` and nothing else, in
any place of a pattern, a group of alternatives too. `FuzzQuote` holds this.

Upstream chose in `abc6998` to document that its `QuoteMeta` does not escape
the comma, issue 78 of upstream. The fork escapes it, because an escape that
is not needed costs nothing, and a missing one changes the match. The
parenthesis keeps text such as `@(` from opening a group (D10).
