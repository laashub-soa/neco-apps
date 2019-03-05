How to run tests
================

1. Prepare dctest environment using `github.com/cybozu-go/neco/dctest`
2. Make snapshot of placemat VMs by `pmctl snapshot save init`
3. Push the current feature branch to GitHub.
4. cd test; make setup && make test
5. cd test; make test-metallb (or test-monitoring, ...)
