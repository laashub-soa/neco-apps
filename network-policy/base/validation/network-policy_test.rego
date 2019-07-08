package kubernetes.test_admission
import data.kubernetes.admission

test_argocd_allowed {
    count(admission.deny) == 0 with input as {
        "request": {
            "kind": {
                "kind": "NetworkPolicy",
                "group": "crd.projectcalico.org"
            },
            "operation": "CREATE",
            "namespace": "argocd",
            "object": {
                "metadata": {
                    "name": "foo"
                },
                "spec": {}
            }
        }
    }
}

test_denied {
    count(admission.deny) > 0 with input as {
        "request": {
            "kind": {
                "kind": "NetworkPolicy",
                "group": "crd.projectcalico.org"
            },
            "operation": "CREATE",
            "namespace": "foo",
            "object": {
                "metadata": {
                    "name": "foo"
                },
                "spec": {
                    "order": 100.0
                }
            }
        }
    }
}

test_large_order_allowed {
    count(admission.deny) == 0 with input as {
        "request": {
            "kind": {
                "kind": "NetworkPolicy",
                "group": "crd.projectcalico.org"
            },
            "operation": "CREATE",
            "namespace": "foo",
            "object": {
                "metadata": {
                    "name": "foo"
                },
                "spec": {
                    "order": 2000.0
                }
            }
        }
    }
}
