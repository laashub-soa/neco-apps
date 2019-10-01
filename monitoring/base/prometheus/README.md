Prometheus
==========

This directory contains the following files:
- K8s manifests for Prometheus
- Configuration files for Prometheus
- Unit tests for the alert rules
    - `$ promtool test rules ./test_rules.yaml`

Notice
------

[Some alert rules](./kube_prometheus_alert_rules.yaml) come from [coreos/kube-prometheus project](https://github.com/coreos/kube-prometheus).
