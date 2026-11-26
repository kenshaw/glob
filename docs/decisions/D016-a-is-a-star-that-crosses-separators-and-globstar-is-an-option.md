# D16. A `**` is a star that crosses separators, and globstar is an option

Status: Proposed.

A `**` matches any sequence of characters, separators included. It does not
consume the separator next to it, so `a/**/b` with the separator `/` matches
`a/x/b` and `a//b`, and not `a/b`. Upstream pins the same in `ed9fb18`, and
issues 58 and 76 of upstream ask about it. The fork matched `a/b` before,
because of the overlap of a prefix and a suffix (D12).

`WithGlobstar` gives the meaning of bash and of other shells: a `**` that is a
whole segment and is followed by a separator also matches no segment. The
parser changes `**/` to `{**/,}`, as upstream does in `ff0540d`. A `**` at the
end of a pattern is not changed.
