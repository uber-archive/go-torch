#!/bin/bash

set -x

export GO15VENDOREXPERIMENT=1

echo "mode: set" > acc.out
FAIL=0

# List all packages to run tests for
PACKAGES=$(glide novendor)
go test -i $PACKAGES
for dir in $PACKAGES;
do
  go test -coverprofile=profile.out $dir || FAIL=$?
  if [ -f profile.out ]
  then
    cat profile.out | grep -v "mode: set" | grep -v "mocks.go" >> acc.out
    rm profile.out
  fi
done

# Failures have incomplete results, so don't send
if [ "$FAIL" -eq 0 ]; then
  goveralls -service=travis-ci -v -coverprofile=acc.out
fi

rm -f acc.out

exit $FAIL
