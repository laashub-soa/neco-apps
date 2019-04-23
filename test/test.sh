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

waitTimeSync() {
  target=$1
  while true; do
    target_time=$(./dcssh cybozu@${target} date +%s)
    host_time=$(date "+%s")
    if [ $(( host_time - target_time )) -lt 5 ]; then
      break
    fi
    sleep 1
  done
}

# Sync with NTP. Nodes restored from snapshot will be out of time.
echo "wait for boot servers..."
for boot in boot-0 boot-1 boot-2 boot-3; do
  for i in $(seq 300); do
    if ./dcssh cybozu@${boot} date 2>/dev/null; then
      break
    fi
    sleep 1
  done
  ./dcssh cybozu@${boot} sudo systemctl restart chronyd.service
  waitTimeSync ${boot}
done

./dcscp sync.sh cybozu@boot-0:
./dcssh cybozu@boot-0 "./sync.sh"

# Restart CKE. Vault token will be expired.
for boot in boot-0 boot-1 boot-2; do
  ./dcssh cybozu@${boot} sudo systemctl start cke.service
done

sudo -E nsenter -t $(pmctl pod show operation | jq .pid) -n sh -c "export PATH=$PATH; $GINKGO"
