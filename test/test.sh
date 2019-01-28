#!/bin/sh

DIR="$1"

pmctl snapshot load init

PLACEMAT_PID=$(echo $(pgrep placemat) | tr " " ",")
operation_pid=$(pgrep -P $PLACEMAT_PID -f operation)
cd ${DIR}
sudo -E nsenter -t ${operation_pid} -n sh -c "export PATH=$PATH; $GINKGO"
exit $?
