# D17. Case folding maps each rune to one rune of its case orbit

Status: Decided.

`WithFold` matches as `strings.EqualFold` compares. The compiler maps each
rune of the text of the pattern, and of the separators, to one rune of its
case folding orbit, and `Match` maps the input the same way. A class matches
a rune when it holds any rune of the orbit of that rune. With the fold option,
`[[:lower:]]` matches `A` too, as `nocasematch` of bash does.

The rune is the smallest lower case rune of the orbit, or the smallest rune
when the orbit holds no lower case rune. So `A` and `a` map to `a`, the Kelvin
sign maps to `k`, and the long s maps to `s`. An input of lower case ASCII text
maps to itself, so `Match` does not copy it.

A rune keeps its length in runes, so the matchers need no change. Its length
in bytes can change, so a caller must not use an offset into the folded
input.
