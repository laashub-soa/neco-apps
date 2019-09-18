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

TBD

cert-manager
------------

Download manifests as follows:

```console
wget https://github.com/jetstack/cert-manager/releases/download/vX.Y.Z/cert-manager.yaml
```

Compare each resource by your eyes.

external-dns
------------

Read the following document and fix manifests as necessary.

https://github.com/kubernetes-incubator/external-dns/blob/vX.Y.Z/docs/tutorials/coredns.md


gatekeeper
----------

Check diffs of open-policy-agent/gatekeeper files as follows:

```console
git clone https://github.com/open-policy-agent/gatekeeper
cd gatekeeper
git diff vA.B.C...vX.Y.Z deploy
```

contour
-------

Check diffs of projectcontour/contour files as follows:

```console
git clone https://github.com/projectcontour/contour
cd contour
git diff vA.B.C...vX.Y.Z examples/contour
```

metallb
-------

Check diffs of danderson/metallb files as follows:

```console
git clone https://github.com/danderson/metallb
cd metallb
git diff vA.B.C...vX.Y.Z manifests
```

prometheus
----------

Check diffs of coreos/kube-prometheus files as follows:

```console
git clone https://github.com/coreos/kube-prometheus
cd kube-prometheus
git diff vA.B.C...vX.Y.Z manifests
```
