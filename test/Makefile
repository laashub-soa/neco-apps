BOOT0 = 10.72.48.0
GINKGO = env GO111MODULE=off $(GOPATH)/bin/ginkgo --failFast -v
SSH_PRIVKEY = $(realpath ./dctest_key)
TEST_ID = $(shell basename $(dir $(GOPATH)))
COMMIT_ID=
export BOOT0 GINKGO SSH_PRIVKEY TEST_ID COMMIT_ID

test:
	./test.sh

setup:
	GO111MODULE=off go get github.com/onsi/ginkgo/ginkgo
	GO111MODULE=off go get -t ./...

# this target expects to be executed after `test` target
test-coil:
	./test.sh ../coil/test
