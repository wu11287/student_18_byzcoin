all: test

PKG_STABLE = gopkg.in/dedis/onet.v2
include $(shell go env GOPATH)/src/github.com/dedis/Coding/bin/Makefile.base

# You can use `test_playground` to run any test or part of cothority
# for more than once in Travis. Change `make test` in .travis.yml
# to `make test_playground`.
test_playground:
	for a in $$( seq 100 ); do \
	  if DEBUG_TIME=true go test -v -race -short > log.txt 2>&1; then \
		  echo Successfully ran at $$(date); \
		else \
		  echo Failed at $$(date); \
			cat log.txt; \
			exit 1; \
		fi; \
	done;
