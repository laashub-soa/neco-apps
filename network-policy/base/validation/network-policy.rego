package networkPolicyOrder

operations = {"CREATE", "UPDATE"}

violation[{"msg": msg, "details": {"order": order}}] {
	operations[input.review.operation]
	matched := {ns | ns := input.parameters.systemNamespaces[i]; ns == input.review.namespace}
	count(matched) == 0
	order := input.review.object.spec.order
	order <= input.parameters.limitOrder
	msg := sprintf("cannot create/update non-system NetworkPolicy with order <= %v", [input.parameters.limitOrder])
}
