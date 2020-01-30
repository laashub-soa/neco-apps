#!/bin/bash -e

if [ -n "$SECRET_GITHUB_TOKEN" ]; then
    GIT_USER=cybozu-neco
    GIT_URL="https://${GIT_USER}:${SECRET_GITHUB_TOKEN}@github.com/cybozu-private/neco-apps-secret"

    if [ "${CIRCLE_BRANCH}" != "release" -a "${CIRCLE_BRANCH}" != "stage" ]; then
        BRANCH="master"
    else
        BRANCH=${CIRCLE_BRANCH}
    fi

    rm -rf ./neco-apps-secret
    git clone -b $BRANCH $GIT_URL neco-apps-secret 2> /dev/null

    kustomize build ./neco-apps-secret/base > expected-secret.yaml
    kustomize build ../secrets/base > current-secret.yaml

elif [ -n "$SECRET_DIR" ]; then
    # By dir
    kustomize build ${SECRET_DIR}/base > expected-secret.yaml
    kustomize build ../secrets/base > current-secret.yaml

else
    echo "Error: Please set env of SECRET_DIR."
    exit 2
fi
