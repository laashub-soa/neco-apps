Build Rook/Ceph OSD on PVC environment using kind
---

# Goal

- kindにおいて、Rook/Ceph OSD on PVCの検証を実施するための環境構築を行う。


export WORKDIR=~/osdpvc
export VG_NAME=ceph_vg
mkdir -p $WORKDIR
truncate --size=70G $WORKDIR/backing_store

sudo losetup -f $WORKDIR/backing_store
losetup -j $WORKDIR/backing_store
sudo vgcreate -y $VG_NAME /dev/loop8
# vgsコマンドで確認 lvs

for i in $(seq 9); do
  sudo lvcreate -y -n ceph_lv_$i -L 6G $VG_NAME
  # ここでハードリンクを作っておく /dev/ceph_lv_xx
done

# ls /dev/ceph_vgで確認可能

kind create cluster --name rook-poc --config kind.yaml
# kind get clustersで確認可能
# kind delete cluster --name rook-pocで削除可能

$ docker exec -it rook-poc-worker /bin/bash
