Neco Ops
========

[![CircleCI](https://circleci.com/gh/cybozu-go/neco-ops.svg?style=svg)](https://circleci.com/gh/cybozu-go/neco-ops)

This repository contains GitOps resources for Neco. It mostly contains Kubernetes deployment resources.

Requirements
------------

- [Kubernetes][]
- [Argo CD][]
- [Kustomize][]

CI/CD
-----

### Overview

![overview](http://www.plantuml.com/plantuml/png/dPJVQzim4CVVzLSSVceWCJJPZyuFWv5ktHZTIblOIw78yiNHR4j6aXl2wFy--P0zBHYQvibdt-VklYVTcGkd3IIN-FIpjO2gb0hH9C3_lR30tEAJn5rmcl32HAsx0f8hwRvsUG9_601uht1SbJL2eb3eXMxjWpgxtye-iDLM-eJx6INg_O_UpvvP56KTBv7yP8MqeTRtBaUZqA7rNphhWgJgmZx6z86GJwRKiBuab7jR54JZo9xD-dDeQxrlK3axVr1hhJQolERj7D29b48B8ivFYbgU5BMQweRNQ5p35Iz3z_47DaR4ZSAYq3kr-3YqaC7xDDZ7yCjNygj3_ld_AswDBaXvRnnzmGGVURy4JvNEAwBowkYoX1vLrADJ9Vr-z0jsZzP9LHBllFsI0ByrOO6SB-32ElMH2LojR_gp--sBp4QX8Uc4WhKqmZ_NtuWLz2RaiBztDHSLUMnlnO6V6ovhSc5lCJRy6COB7tOOCZXFuPfN23UtSn2wQQHZetTXPBbrdX-AUtwPqfc0qwqKa1kkn1OQhkJ6VxLi94CfEhGCLZxAx7rGc2yGMoyMOxx6ZOkaPV0cXRkjte51szME3J-ma3act_eUq9GuAX-PUDmpU9V2B-wytfOR9qLNSnCwc4FRrVMewY3FOin44tfweZXJNhqYp-JQnd-G32vI-ABDZfi2fDfeqt0djM8nD4RIo6Jm8OKVsiaDNGmDr3HEOtI0qv2nlm00)

- Template based or plain K8s manifests are stored in the neco-ops repository.
- To verify changed manifests deploy, CI runs deployment to the `neco-ops` instance in the GCP. See details in `neco-ops` section in this file.
- [Argo CD][] watches changes of this repository, then synchronize(deploy) automatically when new commit detected.
- After deployment process finished, [Argo CD][] sends alert to the [Alertmanager][] where is running on the same cluster. Then it notifies to Slack channel and/or Email address.

### `neco-ops` instance

`neco-ops` is a Google Compute Engine instance for GitOps testing of this repository. It is automatically created and deleted by [CircleCI scheduled workflow](https://circleci.com/docs/2.0/workflows).

The following describes an example of time line based development flow with the Kubernetes cluster in `neco-ops` instance.

1. `neco-ops`: Launch an instance early morning.
2. `neco-ops`: Construct dctest environmmment.
3. `neco-ops`: Download and load virtual machine snapshots from Google Cloud Storage.
4. `neco-ops`: Ready as Kubernetes cluster.
5. `Developer`: Write some code and push commits.
6. `CircleCI`: Deploy commits to `neco-ops` cluster for testing through `Argo CD`.
7. `neco-ops`: Delete an instance at night.
8. `CircleCI`: CI does not work because no `neco-ops` cluster in this time.
9. Go back to 1. tomorrow.

Directory tree
--------------

```console
.
├── argocd-config # Argo CD CRD based app configurations
│   ├── base
│   │   └── monitoring.yaml # CRD yaml for app "monitoring" configuration includes repository URL and path.
│   └── overlays
│       ├── bk
│       ├── prod
│       └── stage
│           ├── kustumization.yaml # Argo CD CRD deployment for stage.
│           └── monitoring.yaml    # overlays for base/monitoring.yaml.
└── monitoring # App "monitoring" deployment manifests.
    ├── base
    │   ├── deployment.yaml    # Plain manifest files of each K8s object name
    │   ├── kustomization.yaml
    │   └── service.yaml
    ├── overlays
    │   ├── dev
    │   ├── prod
    │   └── stage
    │       ├── cpu_count.yaml     # Some tuning
    │       ├── kustomization.yaml
    │       └── proxy.yaml         # NO_PROXY, HTTP_PROXY, HTTPS_PROXY environment variables
    └── test
        └── suite_test.go          # Ginkgo based deployment test
...
```

`argocd-config/overlays/stage/kustomization.yaml`
```yaml
bases: # It includes all applications for stage.
- ../../base
...

patches:
- monitoring.yaml # Argo CD CRD of app "monitoring" for stage.
```

`argocd-config/overlays/stage/monitoring.yaml`
```yaml
# Custom Resource Definition for Argo CD app "monitoring"
spec:
  project: default
  source:
    repoURL: https://github.com/cybozu-go/neco-ops.git
    targetRevision: release         # branch name
    path: monitoring/overlays/stage # Path to Kustomize based app path
    kustomize:
      namePrefix: stage-
  destination:
    server: https://kubernetes.default.svc
    namespace: default
```

`monitoring/overlays/stage/kustomization.yaml`
```yaml
bases:   # It includes all K8s objects for monitoring.
- ../../base
patches: # Patches for stage
- cpu_count.yaml
- proxy.yaml
```

`monitoring/base/kustomization.yaml`
```yaml
resources:   # It includes all K8s objects for monitoring.
- deployment.yaml
- service.yaml
```

Planned Test Flow
-----------------

CI in this repository runs deployment test using `neco-ops` instance. Test resources are in `test/` directory of each `appname`.
Typical test step is:

- Run [Ginkgo][] based deployment test.
    0. Load initialized state of the placemat snapshot by `pmctl snapshot load`.
    1. Login to `neco-opts` instance.
    2. Deploy Argo CD by `kubectl`.
    3. Initialize Argo CD client with `argocd login SERVER --name admin --password xxxxx`.
    4. Deploy Argo CD configuration `argocd-config` by:
        ```console
        argocd app create argocd-config -f https://github.com/cybozu-go/neco-ops --path argocd-config/overlays/stage --dest-namespace=argocd ...
        ````
    5. Deploy `argocd-config` and other apps through Argo CD by `argocd app sync APPNAME`.
    6. Check some status.

License
-------

MIT

[Kubernetes]: https://kubernetes.io/
[Kustomize]: https://github.com/kubernetes-sigs/kustomize
[Argo CD]: https://github.com/argoproj/argo-cd
[Alertmanager]: https://prometheus.io/docs/alerting/alertmanager/
[Ginkgo]: https://github.com/onsi/ginkgo
