How to maintain neco-apps
=========================

## argocd

Check [releases](https://github.com/argoproj/argo-cd/releases) for changes.

Download the upstream manifest as follows:

```console
$ curl -sLf -o argocd/base/upstream/install.yaml https://raw.githubusercontent.com/argoproj/argo-cd/vX.Y.Z/manifests/install.yaml
```

Then check the diffs by `git diff`.

## cert-manager

Check [the upgrading section](https://cert-manager.io/docs/installation/upgrading/) in the official website.

Download manifests and remove `Namespace` resource from it as follows:

```console
$ curl -sLf -o  cert-manager/base/upstream/cert-manager.yaml https://github.com/jetstack/cert-manager/releases/download/vX.Y.Z/cert-manager.yaml
$ vi cert-manager/base/upstream/cert-manager.yaml
  (Remove Namespace resources)
```

## elastic (ECK)

To check diffs between versions, download and compare manifests as follows:

```console
wget https://download.elastic.co/downloads/eck/X.Y.Z/all-in-one.yaml
sed 'N;N;N;N;N;s/apiVersion: v1\nkind: Namespace\nmetadata:\n  name: kube-system//' all-in-one.yaml > all-in-one_nsremoved.yaml
```

## external-dns

Read the following document and fix manifests as necessary.

https://github.com/kubernetes-incubator/external-dns/blob/vX.Y.Z/docs/tutorials/coredns.md

## ingress (Contour & Envoy)

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
  - If the manifest in the upstream is usable as is, use it from `ingress/base/kustomization.yaml`.
  - If the manifest needs modification:
    - If the manifest is for a cluster-wide resource, put a modified version in the `common` directory.
    - If the manifest is for a namespaced resource, put a template in the `template` directory and apply patches.

## metallb

Check [releases](https://github.com/danderson/metallb/releases)

Download manifests and remove `Namespace` resource from it as follows:

```console
$ git clone https://github.com/danderson/metallb
$ cd metallb
$ git checkout vX.Y.Z
$ cp manifests/*.yaml /path/to/neco-apps/metallb/base/upstream
$ vi metallb/base/upstream/metallb.yaml
  (Remove Namespace resources)
```

## monitoring (prometheus, alertmanager, grafana)

There is no official kubernetes manifests for prometheus.
So, check changes in release notes on github for both prometheus and alertmanager and make necessary actions.

## network-policy (Calico)

Check [the release notes](https://docs.projectcalico.org/v3.11/release-notes/).

Download the upstream manifest as follows:

```console
$ curl -sLf -o network-policy/base/calico/upstream/calico-policy-only.yaml https://docs.projectcalico.org/vX.Y/manifests/calico-policy-only.yaml
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
