# D17. Case folding maps each rune to the smallest rune of its case orbit

Status: Proposed.

`WithFold` matches as `strings.EqualFold` compares. The compiler maps each
rune of the text of the pattern, and of the separators, to the smallest rune
of its case folding orbit, and `Match` maps the input the same way. A class
matches a rune when it holds any rune of the orbit of that rune. With the
fold option, `[[:lower:]]` matches `A` too, as `nocasematch` of bash does.

A rune keeps its length in runes, so the matchers need no change. Its length
in bytes can change, so a caller must not use an offset into the folded
input.
