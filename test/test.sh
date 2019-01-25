#!/bin/sh

TARGET="$1"

pmctl snapshot load init

PLACEMAT_PID=$(echo $(pgrep placemat) | tr " " ",")
operation_pid=$(pgrep -P $PLACEMAT_PID -f operation)
sudo -E nsenter -t ${operation_pid} -n sh -c "export PATH=$PATH; $GINKGO" -focus="${TARGET}"
exit $?
