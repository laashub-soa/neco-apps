#!/bin/bash -e

GIT_USER=cybozu-neco

if [ -z "$SECRET_GITHUB_TOKEN" ]; then
    GIT_URL="git@github.com:cybozu-private/neco-apps-secret.git"
else
    GIT_URL="https://${GIT_USER}:${SECRET_GITHUB_TOKEN}@github.com/cybozu-private/neco-apps-secret"
fi

rm -rf ./neco-apps-secret
git clone -b add-secrets $GIT_URL neco-apps-secret 2> /dev/null

kustomize build ./neco-apps-secret/base > expected-secret.yaml
kustomize build ../secrets/base > current-secret.yaml
