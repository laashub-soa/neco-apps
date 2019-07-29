package kubernetes.admission

groupset[role] { role := input.request.userInfo.groups[_] }

deny[msg] {
    not groupset["system:masters"]
    input.request.kind.kind == "RoleBinding" ; input.request.object.roleRef.kind == "ClusterRole"
	msg := "using ClusterRole in RoleBinding is not allowed for this user"
}
