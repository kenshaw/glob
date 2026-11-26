# D12. The backtracking matcher matches what the optimizer cannot, and is the reference of the tests

Status: Proposed. Amended by D20.

`syntax.BacktrackMatcher` matches a tree directly. It tries each way to split
the input, and it remembers each answer for one call, so a call takes at most
cubic time in the length of the input.

1. A tree with an extended group gets it (D10).
2. A tree that the optimizer cannot build gets it, in place of the error
   "can not index on matchers" that `Compile` returned before. The pattern
   `{{{}*}}{{}*}` is one such tree.
3. The tests compare the optimized matchers with it on every vector and in
   `FuzzMatchersAgree`. On 2026-11-26, the survey and the fuzz test found
   these faults of the optimizer, and each is fixed:
   - a prefix and a suffix that overlap, issues 51 and 61 of upstream,
   - the index of a row,
   - a head of zero runes,
   - the index of an every matcher,
   - the segments of a suffix,
   - the overlapping segments of a prefix and a suffix,
   - a compile time that grew exponentially with the number of stars (D18).

   Each input that failed is in `testdata/fuzz/FuzzMatchersAgree`.
