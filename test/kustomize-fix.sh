#!/bin/sh -ex

for dir in ${KUSTOMIZATION_DIRS}; do
    (cd ${dir}; kustomize edit fix)
done
