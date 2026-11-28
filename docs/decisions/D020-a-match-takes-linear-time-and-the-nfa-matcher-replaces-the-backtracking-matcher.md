# D20. A match takes linear time, and the NFA matcher replaces the backtracking matcher

Status: Decided. Amends D12.

On 2026-11-26, a search of random patterns found that a match could take time
of the order of the length of the input to the power of the number of stars.
The pattern `*a*a*a*a*b` took 40 seconds on an input of 400 bytes. The pattern
`{a*,*b,ab}a?b{*a,b*}*{a,*b}` took two minutes on an input of 1,200 bytes. The
tree matcher of the optimizer tries each occurrence of its value, and matches
each side again for each occurrence, so each level of the tree multiplies the
time. The backtracking matcher of D12 took cubic time, and its recursion was
as deep as the input was long, so a long input could end the program with a
stack overflow.

1. `syntax.NFAMatcher` replaces `syntax.BacktrackMatcher`. It reads the input
   once for each node of the tree, and keeps the set of the positions where
   the nodes so far can end, as the simulation of a nondeterministic
   automaton does. Its recursion is as deep as the pattern, never as the
   input. It does what D12 gave the backtracking matcher to do: it matches
   the trees with extended groups and the trees that the optimizer cannot
   build, and it is the reference of the tests.
2. A pattern without extended groups takes time linear in the length of the
   input on the NFA matcher. A group `+(...)`, `*(...)` or `!(...)` whose
   patterns match at most a fixed number of bytes, such as `+(a|aa)` or
   `!(js|ts)`, takes linear time too. Any other such group takes quadratic
   time at worst. The time of a call never grows faster than that.
3. `syntax.Linear` reports whether a tree of optimized matchers takes linear
   time. It knows each matcher, and its answer is conservative. `Node.Match`
   gives the NFA matcher to each tree that is not linear. A pattern keeps the
   optimized matchers when they are linear, and the matcher that a pattern
   gets never changes its answer.
4. The tree matcher has two modes that make more trees linear. When one side
   is a star and the other side matches each longer string too, the tree
   tries only the occurrence of its value that ends first, or that starts
   last. A side that is a star or nothing is answered from the positions of
   the separators, which the tree finds once for each input.
5. `TestWorstCase` matches each shape that the search found on an input of
   64 KiB. `TestAllocs` makes sure that a match does not allocate.
   `FuzzCompile` gives any bytes to `CompileWith`, with each option, and
   compares the answer with the NFA matcher.

The search and the fuzz tests also found these faults, and each is fixed:

- With more than one separator, the index of the first separator was the
  index of the first separator in the list that the input held, and not the
  smallest index.
- The index of the last separator that is not ASCII looked in the wrong part
  of the input, and could panic.
- The width of a byte that is not valid UTF-8 was 3 in the segments of five
  matchers, and the tail of an input counted such a byte wrongly.
- A pattern with a byte that is not valid UTF-8 after a backslash, or in a
  class, was not an error (D8).
- The debug build, with the tag `globdebug`, did not compile.

The NFA matcher costs time on some patterns that the optimized matchers held
before. `[a-z][!a-x]*cat*[h][!b]*eyes*` takes 180 ns in place of 90 ns, because
its tree of optimized matchers is not linear. Open question 6 in `PLAN.md`
asks the maintainers if that cost is acceptable.
