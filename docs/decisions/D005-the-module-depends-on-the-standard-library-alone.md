# D5. The module depends on the standard library alone

Status: Decided.

`go.mod` has no `require` line. A program that imports `glob` gets the
standard library and nothing else. A new dependency needs a decision of the
maintainers. `TestTheModuleHasNoRequire` holds this.

This follows magic D14.
