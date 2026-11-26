# D19. The tests read vectors from the survey of glob syntax

Status: Proposed.

`docs/SYNTAX.md` is a survey of the glob syntax of the Go standard library,
doublestar, bash, POSIX `fnmatch`, minimatch, picomatch, Python, Rust
`globset`, gitignore, editorconfig and upstream. It names what the fork
supports, and why it leaves the rest out.

`testdata/vectors.tsv` holds the test vectors of the survey: a pattern, an
input, the separators and the expected result, with the source of each. Each
vector holds for the fork, or it has the status `differs` and a note that
says why. `TestVectors` runs every vector with the optimized matchers and with
the NFA matcher (D12, D20).
