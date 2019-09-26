#!/bin/sh -e

CI_REMOTE_REPOSITORY="git@github.com:${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}.git"
TAGNAME="$(date +%Y.%m.%d)-$CIRCLE_BUILD_NUM"

git add -u
git commit -m "[ci skip] $TAGNAME"
git tag release-$TAGNAME
git push ${CI_REMOTE_REPOSITORY} stage release-$TAGNAME
