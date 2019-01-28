#!/bin/sh

DIR="$1"
PLACEMAT_PID=$(echo $(pgrep placemat) | tr " " ",")
operation_pid=$(pgrep -P $PLACEMAT_PID -f operation)

if [ -z "${DIR}" ]; then
    # Load snapshot only when no DIR specified
    pmctl snapshot load init
else
    cd ${DIR}
fi

sudo -E nsenter -t ${operation_pid} -n sh -c "export PATH=$PATH; $GINKGO"
exit $?
