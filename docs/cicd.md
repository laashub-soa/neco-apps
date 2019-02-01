CI/CD
=====

Overview
--------

![overview](http://www.plantuml.com/plantuml/png/dPJVQzim4CVVzLSSVceWCJJPZyuFWv5ktHZTIblOIw78yiNHR4j6aXl2wFy--P0zBHYQvibdt-VklYVTcGkd3IIN-FIpjO2gb0hH9C3_lR30tEAJn5rmcl32HAsx0f8hwRvsUG9_601uht1SbJL2eb3eXMxjWpgxtye-iDLM-eJx6INg_O_UpvvP56KTBv7yP8MqeTRtBaUZqA7rNphhWgJgmZx6z86GJwRKiBuab7jR54JZo9xD-dDeQxrlK3axVr1hhJQolERj7D29b48B8ivFYbgU5BMQweRNQ5p35Iz3z_47DaR4ZSAYq3kr-3YqaC7xDDZ7yCjNygj3_ld_AswDBaXvRnnzmGGVURy4JvNEAwBowkYoX1vLrADJ9Vr-z0jsZzP9LHBllFsI0ByrOO6SB-32ElMH2LojR_gp--sBp4QX8Uc4WhKqmZ_NtuWLz2RaiBztDHSLUMnlnO6V6ovhSc5lCJRy6COB7tOOCZXFuPfN23UtSn2wQQHZetTXPBbrdX-AUtwPqfc0qwqKa1kkn1OQhkJ6VxLi94CfEhGCLZxAx7rGc2yGMoyMOxx6ZOkaPV0cXRkjte51szME3J-ma3act_eUq9GuAX-PUDmpU9V2B-wytfOR9qLNSnCwc4FRrVMewY3FOin44tfweZXJNhqYp-JQnd-G32vI-ABDZfi2fDfeqt0djM8nD4RIo6Jm8OKVsiaDNGmDr3HEOtI0qv2nlm00)

- Template based or plain K8s manifests is stored in the neco-ops repository.
- To verify changed manifests deploy, CI runs deployment to the `neco-ops` instance in the GCP. See details in `neco-ops` section in this file.
- [Argo CD][] watches changes of this repository, then synchronize(deploy) automatically when new commit detected.
- After the deployment process finished, [Argo CD][] sends alert to the [Alertmanager][] where is running on the same cluster. Then it notifies to Slack channel and/or Email address.

`neco-ops` instance
-------------------

`neco-ops` is a Google Compute Engine instance for GitOps testing of this repository. It is automatically created and deleted by [CircleCI scheduled workflow](https://circleci.com/docs/2.0/workflows).

The following describes an example of timeline based development flow with the Kubernetes cluster in `neco-ops` instance.

1. `neco-ops`: Launch an instance early morning.
2. `neco-ops`: Construct dctest environment.
3. `neco-ops`: Download and load virtual machine snapshots from Google Cloud Storage.
4. `neco-ops`: Ready as Kubernetes cluster.
5. `Developer`: Write some code and push commits.
6. `CircleCI`: Deploy commits to `neco-ops` cluster for testing through `Argo CD`.
7. `neco-ops`: Delete an instance at night.
8. `CircleCI`: CI does not work because of no `neco-ops` cluster in this time.
9. Go back to 1. tomorrow.

CD of each cluster
------------------

See details of the deployment step in [release.md](release.md).

- stage: watch `argocd-config/overlays/stage` in **master HEAD** branch. All changes of `master` are always deployed to staging cluster.
- bk: TBD
- prod: watch `argocd-config/overlays/prod` in **release HEAD** branch. To deploy changes for a production cluster.
