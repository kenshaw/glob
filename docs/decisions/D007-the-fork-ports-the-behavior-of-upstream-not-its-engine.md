# D7. The fork ports the behavior of upstream, not its engine

Status: Proposed.

In 2026-08, upstream replaced its matching engine with a new one, and released
it as `v1.0.0`. The fork does not take the new engine. It keeps its own
engine, and ports each fix of upstream as a change of behavior, with a test.

The reasons:

1. The fork changed the layout and the API of upstream in 2025, so no commit
   of upstream applies to it as a patch.
2. The tests of upstream, such as `issues_test.go`, name the behavior of each
   fix. The fork runs the same cases in `TestUpstreamIssues`.
3. The fork has a second matcher, the backtracking matcher (D12), so each
   change of the optimized matchers has a reference to agree with.

The ledger records each commit of upstream and what the fork did with it
(D6).
