package kubernetes.test_admission

import data.kubernetes.admission

test_allow_binding_clusterrole_in_clusterrolebinding {
	count(admission.deny) == 0 with input as {"request": {
		"kind": {
			"kind": "ClusterRoleBinding",
		},
        "userInfo": {
            "username": "admin",
            "uid": "014fbff9a07c",
            "groups": ["system:authenticated","system:masters"],
        },
        "object": {
            "roleRef": {
                "apiGroup": "apiGroup: rbac.authorization.k8s.io",
                "kind": "ClusterRole",
                "name": "foo",
            },
        },
	}}
}

test_allow_binding_role_in_rolebinding {
	count(admission.deny) == 0 with input as {"request": {
		"kind": {
			"kind": "RoleBinding",
		},
        "userInfo": {
            "username": "user",
            "uid": "014fbff9a07c",
            "groups": ["system:authenticated","cybozu"],
        },
        "object": {
            "roleRef": {
                "apiGroup": "apiGroup: rbac.authorization.k8s.io",
                "kind": "Role",
                "name": "foo",
            },
        },
	}}
}

test_allow_binding_role_in_clusterrolebinding {
	count(admission.deny) == 0 with input as {"request": {
		"kind": {
			"kind": "ClusterRoleBinding",
		},
        "userInfo": {
            "username": "admin",
            "uid": "014fbff9a07c",
            "groups": ["system:authenticated","system:masters"],
        },
        "object": {
            "roleRef": {
                "apiGroup": "apiGroup: rbac.authorization.k8s.io",
                "kind": "Role",
                "name": "foo",
            },
        },
	}}
}

test_allow_binding_clusterrole_in_rolebinding_by_admin {
	count(admission.deny) == 0 with input as {"request": {
		"kind": {
			"kind": "RoleBinding",
		},
        "userInfo": {
            "username": "admin",
            "uid": "014fbff9a07c",
            "groups": ["system:authenticated","system:masters"],
        },
        "object": {
            "roleRef": {
                "apiGroup": "apiGroup: rbac.authorization.k8s.io",
                "kind": "ClusterRole",
                "name": "foo",
            },
        },
	}}
}


test_deny_binding_clusterrole_in_rolebinding_by_user {
	count(admission.deny) > 0 with input as {"request": {
		"kind": {
			"kind": "RoleBinding",
		},
        "userInfo": {
            "username": "user",
            "uid": "014fbff9a07c",
            "groups": ["system:authenticated","cybozu"],
        },
        "object": {
            "roleRef": {
                "apiGroup": "apiGroup: rbac.authorization.k8s.io",
                "kind": "ClusterRole",
                "name": "foo",
            },
        },
	}}
}
