#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# find all go mod path
# returns an array contains mod path
function find_modules() {
	find . -not \( \
		\( \
		-path './output' \
		-o -path './.git' \
		-o -path '*/third_party/*' \
		-o -path '*/vendor/*' \
		\) -prune \
		\) -name 'go.mod' -print0 | xargs -0 -I {} dirname {}
}

all_modules=$(util::find_modules)

# test all mod
function test() {
	for mod in $all_modules; do
			pushd "$mod" >/dev/null &&
				echo "go test $(sed -n 1p go.mod | cut -d ' ' -f2)" &&
				go test -race -covermode=atomic -coverprofile=coverage.out ./...
			popd >/dev/null || exit
	done
}

# vet all mod
function vet() {
	for mod in $all_modules; do
			pushd "$mod" >/dev/null &&
				echo "go vet $(sed -n 1p go.mod | cut -d ' ' -f2)" &&
				go vet -stdmethods=false ./...
			popd >/dev/null || exit
	done
}

function help() {
	echo "use: test,vet"
}

case $1 in
vet)
	vet
	;;
test)
	test
	;;
*)
	help
	;;
esac
