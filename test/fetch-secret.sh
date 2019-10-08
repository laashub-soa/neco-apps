#!/bin/bash -e

GIT_USER=cybozu-neco

rm -rf ./neco-apps-secret
git clone https://${GIT_USER}:${SECRET_GITHUB_TOKEN}@github.com/cybozu-private/neco-apps-secret neco-apps-secret 2> /dev/null

kustomize build ./neco-apps-secret/base > expected-secret.yaml
kustomize build ../secrets/base > current-secret.yaml
