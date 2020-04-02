# MySQL operator PoC

## Setup

```
# Deploy MySQL Pods
$ kubectl apply -k <neco-apps>/mysql

# Wait all pods become Runnning.
$ kubectl -n app-mysql get pod
NAME          READY   STATUS    RESTARTS   AGE
my-app-db-0   1/1     Running   0          4m1s
my-app-db-1   1/1     Running   0          3m39s
my-app-db-2   1/1     Running   0          3m24s
operator      1/1     Running   0          4m1s

# Setup semi-sync replication
$ kubectl -n app-mysql exec -it operator /etc/mysql/setup/setup.sh
```

## Test

```
$ kubectl -n app-mysql exec -it operator /etc/mysql/setup/do.sh create my-app-db-0
$ kubectl -n app-mysql exec -it operator /etc/mysql/setup/do.sh insert my-app-db-0
$ kubectl -n app-mysql exec -it operator /etc/mysql/setup/do.sh count my-app-db-0
$ kubectl -n app-mysql exec -it operator /etc/mysql/setup/do.sh count my-app-db-1
$ kubectl -n app-mysql exec -it operator /etc/mysql/setup/do.sh count my-app-db-2
```

## Show status

```
$ kubectl -n app-mysql exec -it operator /etc/mysql/setup/status.sh my-app-db-0
$ kubectl -n app-mysql exec -it operator /etc/mysql/setup/status.sh my-app-db-1
$ kubectl -n app-mysql exec -it operator /etc/mysql/setup/status.sh my-app-db-2
```
