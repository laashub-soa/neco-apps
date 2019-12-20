Build Rook/Ceph OSD on PVC environment using kind
---

# Goal

- kindを用いて、Rook/Ceph OSD on PVCの検証を実施するための環境構築を行う。

```
@host
$ export WORKDIR=~/osdpvc
$ export VG_NAME=ceph_vg
$ mkdir -p $WORKDIR
$ truncate --size=70G $WORKDIR/backing_store

$ sudo losetup -f $WORKDIR/backing_store
$ losetup -j $WORKDIR/backing_store

# loop8は要置換
$ sudo vgcreate -y $VG_NAME /dev/loop8

$ for i in $(seq 9); do
  sudo lvcreate -y -n ceph_lv_$i -L 6G $VG_NAME
  # ここでハードリンクを作っておく /dev/ceph_lv_xx
done

# ls /dev/ceph_vg で確認可能

$ kind create cluster --name rook-poc --config kind.yaml
$ kind apply -f pv.yaml

# 確認用の他のterminalなどで実行する
$ kubectl get all -n rook-ceph

# Rook/Cephに関するobjectを適用する
$ kubectl apply -f common.yaml
$ kubectl apply -f operator.yaml
# operatorの準備が待ってから続きを行う
$ kubectl apply -f cluster.yaml
```

## よく使うコマンドメモ
- vgs
- lvs
- kind get clusters
- kind delete cluster --name rook-poc
