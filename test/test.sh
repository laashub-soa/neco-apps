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

EXTERNAL_PID=$(pmctl pod show external | jq .pid)
export EXTERNAL_PID

# Sync with NTP. Nodes restored from snapshot will be out of time.
echo "wait for boot servers..."
for boot in boot-0 boot-1 boot-2 boot-3; do
  for i in $(seq 300); do
    if ./dcssh cybozu@${boot} date 2>/dev/null; then
      break
    fi
    sleep 1
  done
  ./dcssh cybozu@${boot} podenter chrony chronyc makestep
done

./dcssh cybozu@boot-0 "for host in \$(sabactl machines get --role worker | jq -r '.[] | .spec.ipv4[0]'); do ckecli ssh cybozu@\${host} /opt/bin/podenter chrony chrony chronyc makestep; done"

sudo -E nsenter -t $(pmctl pod show operation | jq .pid) -n sh -c "export PATH=$PATH; $GINKGO"
