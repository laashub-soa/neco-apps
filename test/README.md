How to run tests
================

1. Prepare dctest environment using `github.com/cybozu-go/neco/dctest`
2. Make snapshot of placemat VMs by `pmctl snapshot save init`
3. Push the current feature branch to GitHub.
4. Run following commands to setup Argo CD.

    ```console
    cd test
    make setup
    make test
    ```

5. Setup and run tests for apps by `make test-metallb`, `make test-monitoring`, ...  
   If you want to deploy all apps, run `make test-all`.

Running `make test` initializes VMs at the point of the snapshot.
Do it carefully!

Using `argocd`
--------------

`argocd` is a command-line tool to manage Argo CD apps.

Following features are most useful:

* `argocd app list`: list apps and their statuses.
* `argocd app get NAME`: show detailed information of an app.
* `argocd app sync NAME`: immediately synchronize an app with Git repository.
