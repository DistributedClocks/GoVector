#!/usr/bin/env bash

set -e
echo "" > coverage.txt

# Run the coverage test on every single subdirectory
# except the ones that are in the vendor package.
# Currently GoVector has no vendor packages but this
# code is left in as it is vendor-proof.
for d in $(go list ./... | grep -v vendor); do
        go test -race -coverprofile=profile.out -covermode=atomic $d
        if [ -f profile.out ]; then
                cat profile.out >> coverage.txt
                rm profile.out
        fi
done

