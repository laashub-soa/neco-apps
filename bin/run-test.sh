#!/bin/sh

. $(dirname $0)/env

cat >run.sh <<EOF
#!/bin/sh -ex
# Run test
GOPATH=\$HOME/${TEST_DIR}/go
export GOPATH
PATH=/usr/local/go/bin:\$GOPATH/bin:\$PATH
export PATH
git clone https://github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME} \$HOME/${TEST_DIR}/go/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
cd \$HOME/${TEST_DIR}/go/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
git checkout -qf ${CIRCLE_SHA1}
cd test
export GO111MODULE=on
make setup
make COMMIT_ID=${CIRCLE_SHA1} test

targets="$(git diff origin/master ${CIRCLE_SHA1} --name-only | cut -d '/' -f 1 | uniq)"
for target in ${targets}; do
    if test -f ../${target}/test/suite_test.go; then
        echo "Run test-${target}"
        make COMMIT_ID=${CIRCLE_SHA1} test-${target}
    fi
done
EOF
chmod +x run.sh

$GCLOUD compute ssh --zone=${ZONE} cybozu@${INSTANCE_NAME} --command="mkdir -p /home/cybozu/${TEST_DIR}"
$GCLOUD compute scp --zone=${ZONE} run.sh cybozu@${INSTANCE_NAME}:${TEST_DIR}
$GCLOUD compute ssh --zone=${ZONE} cybozu@${INSTANCE_NAME} --command="/home/cybozu/${TEST_DIR}/run.sh"
RET=$?
#$GCLOUD compute ssh --zone=${ZONE} cybozu@${INSTANCE_NAME} --command="sudo rm -rf /home/cybozu/${TEST_DIR}"

exit $RET
