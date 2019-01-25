#!/bin/bash
argocd login argocd-server --insecure --username admin --password password
argocd app create ${APPNAME} \
--repo ${REPO} \
--path ${APP_PATH} \
--dest-namespace ${NAMESPACE} \
--dest-server https://kubernetes.default.svc \
--sync-policy automated
