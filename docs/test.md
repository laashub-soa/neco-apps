Neco Ops Test
=============

neco-ops has `test/` directories on each application directory. e.g. `ingress/test`.
Argo CD setup automatic deployment by [../test](../test) as test.

Synopsis
--------

[`../test/Makefile`](../test/Makefile) runs test for neco-ops on the running dctest.

* `make setup`

    Install test required components.

* `make clean`

    Delete generated files.

* `make kustomize-check`

     Check syntax of the Kubernetes manifests using `kustomize check`

* `make test`

    Deploy Argo CD as test. It is required other test targets which starts with `test-`.

* `make test-APPNAME`

    Deploy APPNAME as test.

* `make test-all`

    Deploy all applications as test.

Options
-------

### `../test/account.json` file

External DNS in Argo CD app `ingress` requires Google Application Credentials in JSON file.
neco-ops test runs `kubectl create secrets .... --from-file=account.json` to register `Secret` for External DNS.
To run `ingress` test, put your account.json of the Google Cloud service account which has a role `roles/dns.admin`.
See details of the role at https://cloud.google.com/iam/docs/understanding-roles#dns-roles
