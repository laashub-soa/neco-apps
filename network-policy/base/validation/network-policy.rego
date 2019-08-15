package networkPolicyOrder

operations = {"CREATE", "UPDATE"}

violation[{"msg": msg, "details": {"order": order}}] {
	operations[input.review.operation]
	not input.parameters.systemNamespaces[input.review.namespace]
	order := input.review.object.spec.order
	order <= input.parameters.limitOrder
	msg := sprintf("cannot create/update non-system NetworkPolicy with order <= %v", [input.parameters.limitOrder])
}
