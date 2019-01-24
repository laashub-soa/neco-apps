BOOT0 = 10.72.48.0
GINKGO = env GO111MODULE=off $(GOPATH)/bin/ginkgo --failFast -v
SSH_PRIVKEY = $(realpath ./dctest_key)
export BOOT0 GINKGO SSH_PRIVKEY

test:
	./test.sh

setup:
	GO111MODULE=off go get github.com/onsi/ginkgo/ginkgo
	go get -t ./...
