# D18. The search of the optimizer has a budget

Status: Decided.

The optimizer tries every way to glue neighbors in a pattern, such as `*` and
`?` into one matcher. The cost of that search grows exponentially with the
number of matchers, and a pattern of 20 stars took 34 seconds to compile on
2026-11-26.

1. The compiler joins each run of stars into one star before the search. A run
   that holds a `**` is a `**`.
2. The search stops after `minimizeBudget` steps, 4,096, and keeps the best
   result so far. Every result matches the same strings, so the budget changes
   the speed of a match and never its answer.
