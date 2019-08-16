package networkPolicyOrder.test_violation

import data.networkPolicyOrder

test_allow_creating_network_policy_in_system_namespaces {
	count(networkPolicyOrder.violation) == 0 with input as {
		"review": {
			"kind": {
				"kind": "NetworkPolicy",
				"group": "crd.projectcalico.org",
			},
			"operation": "CREATE",
			"namespace": "internet-egress",
			"object": {
				"metadata": {"name": "foo"},
				"spec": {},
			},
		},
		"parameters": {
			"systemNamespaces": ["kube-system", "argocd"],
			"limitOrder": 1000,
		},
	}
}

test_deny_creating_high_priority_network_policy {
	count(networkPolicyOrder.violation) > 0 with input as {
		"review": {
			"kind": {
				"kind": "NetworkPolicy",
				"group": "crd.projectcalico.org",
				"version": "v1",
			},
			"operation": "CREATE",
			"namespace": "foo",
			"object": {
				"metadata": {"name": "foo"},
				"spec": {"order": 100},
			},
		},
		"parameters": {
			"systemNamespaces": ["kube-system", "argocd"],
			"limitOrder": 1000,
		},
	}
}

test_deny_creating_high_priority_network_policy_with_old_version {
	count(networkPolicyOrder.violation) > 0 with input as {
		"review": {
			"kind": {
				"kind": "NetworkPolicy",
				"group": "crd.projectcalico.org",
				"version": "v1beta1",
			},
			"operation": "CREATE",
			"namespace": "foo",
			"object": {
				"metadata": {"name": "foo"},
				"spec": {"order": 100},
			},
		},
		"parameters": {
			"systemNamespaces": ["kube-system", "argocd"],
			"limitOrder": 1000,
		},
	}
}

test_deny_updating_high_priority_network_policy {
	count(networkPolicyOrder.violation) > 0 with input as {
		"review": {
			"kind": {
				"kind": "NetworkPolicy",
				"group": "crd.projectcalico.org",
			},
			"operation": "UPDATE",
			"namespace": "foo",
			"object": {
				"metadata": {"name": "foo"},
				"spec": {"order": 100},
			},
		},
		"parameters": {
			"systemNamespaces": ["kube-system", "argocd"],
			"limitOrder": 1000,
		},
	}
}

test_allow_creating_low_priority_network_policy {
	count(networkPolicyOrder.violation) == 0 with input as {
		"review": {
			"kind": {
				"kind": "NetworkPolicy",
				"group": "crd.projectcalico.org",
			},
			"operation": "CREATE",
			"namespace": "foo",
			"object": {
				"metadata": {"name": "foo"},
				"spec": {"order": 2000},
			},
		},
		"parameters": {
			"systemNamespaces": ["kube-system", "argocd"],
			"limitOrder": 1000,
		},
	}
}
