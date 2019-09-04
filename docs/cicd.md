CI/CD
=====

Overview
--------

![overview](http://www.plantuml.com/plantuml/svg/fPJVQzim4CVVzLSSUMeXCRJTZyuFeoFTkZ6wbRQmbq9Hv8kZMPQCT6u8e__xv9ljQY6mj7loMVfyxuTqfxD0qbDR6o4LEG-JStputWH0MsgBw2SWdtu6vXeVDAxxJT__26LSMy3aGjFdTZ61Nm80uguYQKk3CB6et4msJRYp1xShtIaR5tJqk3baJnrmtm7YKIIwc6693B3LE-wZVMqNw2qIXhZI1kgJgax3VKflfVB1bmxcvupAQAlY3pso1uVTIJJ6RMgq5APtTkxiKfUNifa2aifOMZ1nz2BLyOjK9xkgaGOzrTBAigy-NH2Z8YqKPeLRszdxeOGStcQzlGz__4p-P9j__FkA6-yAphmpzhvWXlUyNuPtXLvKJ2sglSCokbVGYEuA6IUa6x5R3AHjoJn5-ry9nB7X4N8DnG3K3rIddC35_AexkvynoE6OQE9qAuEvzYf-vr-OLVodz331DqQgYdT2PwN2ZxNKXhUmiuGOdeRnmiSXXXoEChYK5SBLzHIKgsjDucbxvdMvegWOmaV1SSQd0kGQhM3XfLKhY8sibt4rY94SWdKLvd0ILNuJHNs7WLh5T37b3IufJIw7LndShDmQF8RMa1XUiT5rWhwEPQ0lCKb-eDAUp-5D1ZyanPGRwKchraZV5p65fZLcJ6pRqOnTZNtwFvuILulg6OvsJc_waEHmci4dvzVI5w3jqlbQadPMfD2evCx9uLq6tnpf9TyEzzLkdBjf2-TU4sTeYxOslm40)

- Template based or plain K8s manifests is stored in the neco-apps repository.
- To verify changed manifests deploy, CI creates some GCP instances and runs bootstrap/upgrade tests on those instances.
- [Argo CD][] watches changes of this repository, then synchronize(deploy) automatically when new commit detected.
- After the deployment process finished, [Argo CD][] sends alert to the [Alertmanager][] where is running on the same cluster. Then it notifies to Slack channel and/or Email address.

GCP instance
------------

This repository uses Google Compute Engine instance for GitOps testing. The instances are automatically created and deleted by [CircleCI][] depending on the job contents.

The GCP instance name is `neco-apps-<CircleCI Build Number>`. If the job succeeds, the corresponding GCP instance will be deleted immediately. When the job failed, the GCP instance remains for a while.

CircleCI Workflow
-----------------

This repository has 2 CircleCI workflows, `main` and `production-release`.

### `main` workflow

`main` workflow is used for testing feature branch of `neco-apps`. This is consists of the following 5 jobs.

| job name          | description                                         | target branch                              |
| ----------------- | --------------------------------------------------- | ------------------------------------------ |
| `test`            | Syntax check for go lang                            | all branches                               |
| `bootstrap-1`     | Bootstrap test                                      | all branches except `master` and `release` |
| `bootstrap-2`     | Bootstrap test with `neco`'s feature branch         | `with-neco-branch-*` branch                |
| `upgrade-master`  | Upgrade test from `master` branch (staging env)     | all branches except `master` and `release` |
| `upgrade-release` | Upgrade test from `release` branch (production env) | all branches except `master` and `release` |

`bootstrap-1`, `upgrade-master` and `upgrade-release` are tested with `neco`'s `release` branch.

`bootstrap-2` is tested with `neco`'s feature branch which extracted from `neco-apps`'s branch name.
For example, when `with-neco-branch-foo-bar` branch of `neco-apps`, it's tested with `foo-bar` branch of `neco`.

### `production-release` workflow

`production-release` workflow is used for releasing `neco-apps` to a production environment.
This workflow is executed only when a `release-*` tag is created. And it creates a pull request for the release.

CD of each cluster
------------------

See details of the deployment step in [deploy.md](deploy.md).

- stage: watch `argocd-config/overlays/stage` in **master HEAD** branch. All changes of `master` are always deployed to staging cluster.
- bk: TBD
- prod: watch `argocd-config/overlays/prod` in **release HEAD** branch. To deploy changes for a production cluster.

[Argo CD]: https://github.com/argoproj/argo-cd
[Alertmanager]: https://prometheus.io/docs/alerting/alertmanager/
[CircleCI]: https://circleci.com/
