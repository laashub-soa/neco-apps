Guide for developers
====================

Adding a new application
------------------------

まずアプリ名のディレクトリを作り、その下のディレクトリ構造を作る。
`overlays` 下には環境名(`gcp`, `stage0`, `tokyo0`)のディレクトリを掘る。
(差分がない環境については `overlays` 下は作成不要。)

```console
$ APP_NAME=foo
$ mkdir -p ${APP_NAME}/base
$ mkdir -p ${APP_NAME}/overlays/gcp
$ mkdir -p ${APP_NAME}/overlays/stage0
$ mkdir -p ${APP_NAME}/overlays/tokyo0
```

### 共通部分

アプリのマニフェストのうち、環境に依らない部分を base に置く。
overlays により補完するので、ここに置いたファイルは一部が欠けたもので
よいこともある。
(パッチストラテジーに依る。)

リソースごとに別ファイルとすると管理しやすい。

!!! Note
    [Argo CD][] による検出の都合上、YAML ファイルの拡張子は `.yml` ではなく
    `.yaml` とする。

```console
$ edit ${APP_NAME}/base/deployment.yaml
$ edit ${APP_NAME}/base/service.yaml
```

これらを [Kustomize][] で束ねる。

!!! Note
    単にファイルを束ねるだけではなく、ConfigMap の YAML ファイルを(ファイルに
    手でテキストブロックを埋め込むのではなく)ファイルから作成するとか、
    各リソースに共通ラベルを付与するとか、いろいろできる。

```console
$ edit ${APP_NAME}/base/kustomization.yaml
```

```yaml
apiVersion: v1beta1
kind: Kustomization
resources:
- deployment.yaml
- service.yaml
```

### 環境依存部分

環境(`gcp`, `stage0`, `tokyo0`)ごとにファイルを作る。
環境による部分のパッチを置く。

```console
$ ENVIRONMENT=stage
$ edit ${APP_NAME}/overlays/${ENVIRONMENT}/deployment_replicas.yaml
```

[Kustomize][] で base を指定しつつパッチを束ねる。

```
$ edit ${APP_NAME}/overlays/${ENVIRONMENT}/kustomization.yaml
```

```yaml
apiVersion: v1beta1
kind: Kustomization
bases:
- ../../base
patches:
- deployment_replicas.yaml
```

環境による部分がない場合は、`${APP_NAME}/overlays/${ENVIRONMENT}` の
ディレクトリごと作成不要である。

### マニフェストの合成確認

環境ごとにマニフェストが正しく合成されるか、確認する。

```console
$ (cd test; make setup)
$ kustomize build ${APP_NAME}/base > /tmp/app.yaml
$ kustomize build ${APP_NAME}/overlays/${ENVIRONMENT} > /tmp/app.yaml
(各環境について /tmp/app.yaml の中身を確認する)
```

!!! note
    合成されたマニフェストを目で確認するのは難しい。
    少なくとも合成そのものが成功することは確認したい。

    GCP 環境用の合成結果を手元の placemat 環境で試すには、まずマニフェストの
    文法等のチェックとして `kubectl apply -f app.yaml --dry-run` が使える。
    さらに `kubectl apply -f app.yaml` で適用すると、実際に使えるマニフェストで
    あるかどうかが分かる。
    この際に、用いられている namespace が単一で非明示的なら
    `kubectl apply --namespace=NS` で指定できるが、複数・明示的な場合は
    非明示的な方を `kubectl config set-context CONTEXT --namespace=NS` で
    コンテキストに設定する必要がある。

    GCP 環境について確認が取れれば、それ以外の環境については、環境間で
    合成結果の diff を取って意図した差分となっているか確認することができる。

### Argo CD への登録

ここまでに作成したリソースを、[Argo CD][] の対象となるアプリケーションとして
登録する。
`argocd-config` ディレクトリ下を変更する。

base および環境ごとに CRD Application とその差分の YAML ファイルを作成する。
差分がない環境については overlays は不要。

```console
$ edit argocd-config/base/${APP_NAME}.yaml
$ edit argocd-config/overlays/${ENVIRONMENT}/${APP_NAME}.yaml
```

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
(中身については省略; いじるのは以下のあたり
  - metadata/name
  - spec/source/path
  - spec/destination/namespace
  - spec/syncPolicy/automated
)
```

追加したファイルを kustomization.yaml に記載する。

```console
$ edit argocd-config/base/kustomization.yaml
$ edit argocd-config/overlays/${ENVIRONMENT}/kustomization.yaml
```

Argo CD 用のマニフェストの合成についても確認できる。

```console
$ kustomize build argocd-config/base
$ kustomize build argocd-config/overlays/${ENVIRONMENT}
```

### テスト追加

そのアプリのデプロイのテスト、およびデプロイされたアプリの動作のテストを
追加する。

```console
$ edit ${APP_NAME}/test/suite_test.go
$ edit ${APP_NAME}/test/...  # デプロイのテストや動作のテスト
```

新しく追加したアプリを`DEPLOY_ORDER`に追加する。`make test-all`実行時のデプロイ順序になっているため、順番を考慮する。
例えば、`Type: LoadBalancer`の`Kind: Service`をデプロイする場合は、`metallb`よりも後にする必要がある。

```console
$ edit test/Makefile
```

例
```makefile
# Deployment order for make test-all
DEPLOY_ORDER = \
        metallb \
        teleport \
        policyagent \
        rook
```

### 適用

#### Staging 環境への適用

Staging 環境については、変更が master ブランチにマージされれば自動的に
適用される。

#### Prod 環境への適用

Prod 環境については、master が Staging 環境で検証されたのちにまず以下の
操作を行う。

```console
# Prod 用手順
$ git checkout master
$ git tag release-$(date +%Y.%m.%d)-(同日の通し番号)
$ git push --tags
```

これにより release ブランチへのプルリクエストが作成される。
これをマージすると自動的に適用される。


アプリケーションの変更
----------------------

追加と同様に頑張る。

!!! warning
    自動同期されないアプリケーションについては手動による同期をとる必要がある。
    現時点では全てのアプリケーションが自動同期の設定となっている。

    ```console
    # syncの状態を確認する。
    $ argocd app list
    
    # 以下でシンクされていないものをシンクさせる。
    $ argocd app sync APP_NAME
    ```


トラブルシューティング
----------------------

`ckecli kubernetes issue` は済んでいるものとする。

### ログイン

`argocd` コマンドを使うためには Argo CD にログインする必要がある。
これは 1 つの環境について一度行えばよい。

パスワードは[秘密のメモ](https://bozuman.cybozu.com/k/33029/)を参照のこと。

```console
$ NODE_ADDRESS=$(kubectl get nodes -o json | jq -r '.items[0].status.addresses[] | select(.type == "InternalIP") | .address')
$ echo ${NODE_ADDRESS}
$ NODE_PORT=$(kubectl get svc/argocd-server -n argocd -o json | jq -r '.spec.ports[] | select(.name == "http") | .nodePort')
$ echo ${NODE_PORT}
$ argocd login ${NODE_ADDRESS}:${NODE_PORT} --insecure --username admin
Password: (パスワードを入力する)
```

### AppOutOfSync

いずれかのアプリケーションの同期がとれていない状態である。
アラートを発した環境で、まずどのアプリケーションに問題があるのかを調べる。

```console
$ argocd app list
NAME           CLUSTER                         NAMESPACE       PROJECT  STATUS     HEALTH       SYNCPOLICY  CONDITIONS
argocd-config  https://kubernetes.default.svc  argocd          default  Synced     Healthy      <none>      <none>
metallb        https://kubernetes.default.svc  metallb-system  default  Synced     Progressing  <none>      <none>
monitoring     https://kubernetes.default.svc  monitoring      default  OutOfSync  Missing      <none>      <none>
```

`STATUS`, `HEALTH`, `SYNCPOLICY` の欄が重要。
`STATUS: OutOfSync` となっているアプリケーションがまず怪しい。
ほかには、あるアプリケーションが `STATUS: Synced` ではあるものの
`HEALTH: Healthy` ではないために、次のアプリケーションの処理に進めず
とばっちりで `STATUS: OutOfSync` となることもあるかもしれない。
その場合は `STATUS: Synced` であっても `HEALTH` に異常のあるものが怪しい。

怪しいアプリケーションについて、手動での同期を試したりリソースの情報を
見たりする。

```console
$ argocd app sync monitoring
$ argocd app get monitoring
```

`Target`, `Path` の指す先や `Sync Status`, `Sync Revision` に現れる
コミット ID (あるいはブランチ名や `HEAD`)が正しいか確認する。
また、`STATUS` や `HEALTH` が怪しいリソースがあったらその情報を見る。

```console
$ kubectl describe -n monitoring deployment alertmanager
```

[Argo CD]: https://github.com/argoproj/argo-cd
[Kustomize]: https://github.com/kubernetes-sigs/kustomize
