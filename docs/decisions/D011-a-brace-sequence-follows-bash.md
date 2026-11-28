# D11. A brace sequence follows bash

Status: Decided.

`{lo..hi}` and `{lo..hi..step}` are sequences, as in bash: `{1..10}`,
`{01..10..2}`, `{-2..2}`, `{a..e}` and `{z..a..2}`.

1. The parser expands a sequence to the alternatives of `{1,2,3}`. A sequence
   with more than `syntax.MaxSequence` items, 10,000, is an error.
2. A number with a leading zero pads each item to the width of the wider end,
   sign included.
3. The step is the absolute value of the third part, and 0 counts as 1. The
   sequence runs from lo to hi in either direction.
4. Only the full body of the braces makes a sequence. `{1..}` and `{a..9}` are
   alternatives that hold one pattern, as before.
