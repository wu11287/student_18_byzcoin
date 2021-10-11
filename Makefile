all: test_fmt test_lint test_local

# gopkg fits all v1.1, v1.2, ... in v1
# gopath=$(shell go env GOPATH)
# include $(gopath)/src/dedis/Coding/bin/Makefile.base
include  $(GOPATH)/src/student_18_byzcoin/Makefile.base

# You can use `test_playground` to run any test or part of cothority
# for more than once in Travis. Change `make test` in .travis.yml
# to `make test_playground`.
test_playground:
	# cd skipchain; \
	cd /root/go/src/student_18_byzcoin/omniledger/vendor/gopkg.in/dedis/cothority.v2/skipchain; \
	for a in $$( seq 100 ); do \
	  go test -v -race -short || exit 1 ; \
	done;

proto:
	awk -f proto.awk struct.go > external/proto/sicpa.proto
	cd external; make

test_simple: test_fmt test_lint test_goveralls
