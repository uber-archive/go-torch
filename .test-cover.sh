#!/bin/bash

# run on CI service w/something like:
#
# go get github.com/axw/gocov/gocov
# go get github.com/mattn/goveralls
# COVERALLS="-service drone.io -repotoken $COVERALLS_TOKEN" ./test-coverage.sh
#

echo "mode: set" > acc.out
fail=0

# Standard go tooling behavior is to ignore dirs with leading underscors
for dir in $(find . -maxdepth 10 -not -path './.git*' -not -path '*/_*' -type d);
do
  if ls $dir/*.go &> /dev/null; then
    go test -coverprofile=profile.out $dir || fail=1
    if [ -f profile.out ]
    then
      cat profile.out | grep -v "mode: set" >> acc.out
      rm profile.out
    fi
  fi
done

# Failures have incomplete results, so don't send
if [ -n "$COVERALLS" ] && [ "$fail" -eq 0 ]; then
  goveralls -v -coverprofile=acc.out $COVERALLS
fi

rm -f acc.out

exit $fail
