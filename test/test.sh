#!/bin/sh

DIR="$1"

if [ -z "${DIR}" ]; then
    # Load snapshot only when no DIR specified
    pmctl snapshot load init
else
    cd ${DIR}
fi

PLACEMAT_PID=$(echo $(pgrep placemat) | tr " " ",")

while true; do
    if pmctl pod show operation >/dev/null 2>&1; then break; fi
    if ! ps -p $PLACEMAT_PID > /dev/null; then
        echo "FAIL: placemat is no longer working."
        exit 1;
    fi
    echo "preparing placemat..."
    sleep 1
done

sudo -E nsenter -t $(pmctl pod show operation | jq .pid) -n sh -c "export PATH=$PATH; $GINKGO"
