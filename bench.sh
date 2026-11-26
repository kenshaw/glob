#!/bin/bash
#
# Compares the benchmarks of the current branch with the benchmarks of a git
# revision. The second argument is the -bench regular expression:
#
#	./bench.sh v0.1.1
#	./bench.sh main 'Match'
#
# The script writes the results to *.bench files in the current folder, and
# compares them with benchstat. Install it with:
#
#	go install golang.org/x/perf/cmd/benchstat@latest
#
# The script checks out each revision, so commit or stash your changes first.

set -eu

prev=$1
what=${2:-.}
curr=$(git rev-parse --abbrev-ref HEAD)
rnd=$(head -c4 </dev/urandom | xxd -p)

file() {
	echo "$rnd-$1.bench" | tr "/" "_"
}

bench() {
	local rev=$1
	local out
	out=$(file "$rev")
	if [[ -e "$out" ]]; then
		echo "$out exists"
		return
	fi
	git checkout -q "$rev"
	echo -n "writing $out... "
	go test ./... -run=NONE -benchmem -bench="$what" >"$out"
	echo "done"
	git checkout -q "$curr"
	sleep 5
}

bench "$prev"
bench "$curr"

benchstat "$(file "$prev")" "$(file "$curr")"
