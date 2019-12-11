How to maintain neco-apps
=========================

argocd
------

Upstream `install.yaml` is generated with kustomize as follows:

```console
kustomize build "${SRCROOT}/manifests/cluster-install" >> "${SRCROOT}/manifests/install.yaml"
```

So, check diffs of argo-cd/manifests files as follows:

```console
git clone https://github.com/argoproj/argo-cd
cd argocd-cd
git diff vA.B.C...vX.Y.Z manifests
```

elastic cloud on Kubernetes
---------------------------

To check diffs between versions, download and compare manifests as follows:

```console
wget https://download.elastic.co/downloads/eck/X.Y.Z/all-in-one.yaml
sed 'N;N;N;N;N;s/apiVersion: v1\nkind: Namespace\nmetadata:\n  name: kube-system//' all-in-one.yaml > all-in-one_nsremoved.yaml
```

cert-manager
------------

Download manifests and remove `Namespace` resource from it as follows:

```console
wget https://github.com/jetstack/cert-manager/releases/download/vX.Y.Z/cert-manager.yaml
sed 'N;N;N;N;N;s/apiVersion: v1\nkind: Namespace\nmetadata:\n  name: cert-manager//' cert-manager.yaml > cert-manager_nsremoved.yaml
```

Note that `cert-manager_nsremoved.yaml` is used for input of `kustomize build`.

external-dns
------------

Read the following document and fix manifests as necessary.

https://github.com/kubernetes-incubator/external-dns/blob/vX.Y.Z/docs/tutorials/coredns.md


contour
-------

Check diffs of projectcontour/contour files as follows:

```console
$ git clone https://github.com/projectcontour/contour
$ cd contour
$ git diff vA.B.C...vX.Y.Z examples/contour
```

Then, import YAML manifests as follows:

```console
$ git checkout vX.Y.Z
$ cp examples/contour/*.yaml /path/to/neco-apps/ingress/base/contour/upstream
```

Note that:
- We do not use contour's certificate issuance feature, but use cert-manager to issue certificates required for gRPC.
- We change Envoy manifest from DaemonSet to Deployment.
- Not all manifests inherit the upstream. Please check `kustomization.yaml` which manifest inherits or not.

metallb
-------

Check diffs of danderson/metallb files as follows:

```console
git clone https://github.com/danderson/metallb
cd metallb
git diff vA.B.C...vX.Y.Z manifests
```

prometheus, alertmanager, grafana
---------------------------------

There is no official kubernetes manifests for prometheus.
So, check changes in release notes on github for both prometheus and alertmanager and make necessary actions.


calico
------

To check diffs between minor versions, download and compare manifests as follows:

```console
wget https://docs.projectcalico.org/vX.Y/manifests/calico-policy-only.yaml -O vX.Y.yaml
wget https://docs.projectcalico.org/vA.B/manifests/calico-policy-only.yaml -O vA.B.yaml
diff -u vX.Y.yaml vA.B.yaml
```

teleport
--------

There is no official kubernetes manifests actively maintained for teleport.
So, check changes in release notes on github.


topolvm
-------

Check diffs of cybozu-go/topolvm files as follows:

```console
git clone https://github.com/cybozu-go/topolvm
cd topolvm
git diff vA.B.C...vX.Y.Z deploy
```
