BOOT0 = 10.72.48.0
GINKGO = env GO111MODULE=on GOFLAGS=-mod=vendor $(GOPATH)/bin/ginkgo --failFast -v
SSH_PRIVKEY = $(realpath ./dctest_key)
TEST_ID = $(shell basename $(dir $(GOPATH)))
COMMIT_ID = $(shell git rev-parse --abbrev-ref HEAD)
KUSTOMIZATION_DIRS = $(shell find ../ -name "kustomization.yaml" -exec dirname {} \;)
export BOOT0 GINKGO SSH_PRIVKEY TEST_ID COMMIT_ID KUSTOMIZATION_DIRS

### for Go
GOFLAGS = -mod=vendor
export GOFLAGS

# Follow Argo CD installed kustomize version
# https://github.com/cybozu/neco-containers/blob/master/argocd/Dockerfile#L32
KUSTOMIZE_VERSION = 2.0.3

# Deployment order for make test-all
DEPLOY_ORDER = \
	metallb

install.yaml: $(shell find ../argocd/base)
	kustomize build ../argocd/base/ > install.yaml

test: install.yaml
	./test.sh

kustomize-check:
	./kustomize-check.sh

# NOTE: Enable this target when Argo CD uses kustomize 2.0.0
#kustomize-fix:
#	./kustomize-fix.sh

setup:
	curl -sSLf -O https://github.com/kubernetes-sigs/kustomize/releases/download/v$(KUSTOMIZE_VERSION)/kustomize_$(KUSTOMIZE_VERSION)_linux_amd64
	sudo mv kustomize_$(KUSTOMIZE_VERSION)_linux_amd64 /usr/local/bin/kustomize
	chmod +x /usr/local/bin/kustomize
	go install github.com/onsi/ginkgo/ginkgo

save:
	./save.sh

# this target expects to be executed after `test` target
test-%: install.yaml
	./test.sh ../$*/test

test-all:
	$(MAKE) test
	for target in $(DEPLOY_ORDER); do \
		$(MAKE) test-$${target}; \
	done
	# monitoring test is always last step for checking all deployed metrics
	$(MAKE) test-monitoring

clean:
	rm -f install.yaml

.PHONY:	test kustomize-check setup save test-monitoring test-metallb test-all clean
