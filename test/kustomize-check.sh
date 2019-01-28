#!/bin/sh -ex

for yaml in ${KUSTOMIZATION_YAMLS}; do
    kustomize build ${yaml} >/dev/null
done
