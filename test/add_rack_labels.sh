#! /bin/sh

kubectl label nodes kindtest-worker rack=1
kubectl label nodes kindtest-worker2 rack=2
kubectl label nodes kindtest-worker3 rack=3
