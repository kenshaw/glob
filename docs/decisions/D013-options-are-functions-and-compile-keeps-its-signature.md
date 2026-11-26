# D13. Options are functions, and Compile keeps its signature

Status: Proposed.

`Compile(pattern, separators...)` keeps its signature, so no caller changes.
`CompileWith(pattern, opts...)` takes options:

- `WithSeparators(separators...)`, which `Compile` passes.
- `WithFold()`, to match without regard to case, issue 37 of upstream.
- `WithGlobstar()`, to make a `**` segment match no segment too (D16).

`New(opts...)` makes an empty glob with options, and
`Glob.UnmarshalText` compiles with them. `New()` with no option works as
before. `Glob.Separators` returns the separators, as upstream added in
`a82038c`. `Glob.MatchBytes` matches a byte slice without a copy, issue 32
and pull request 81 of upstream.
