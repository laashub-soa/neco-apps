package test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testAdmission() {
	It("should validate Calico NetworkPolicy", func() {
		networkPolicyYAML := `
apiVersion: crd.projectcalico.org/v1
kind: NetworkPolicy
metadata:
  name: admission-test
  namespace: default
spec:
  order: 100.0
  selector: app.kubernetes.io/name == 'hoge'
  types:
  - Ingress
  ingress:
  - action: Allow
    protocol: TCP
    destination:
      ports:
      - 8000
`
		_, stderr, err := ExecAtWithInput(boot0, []byte(networkPolicyYAML), "kubectl", "apply", "-f", "-")
		Expect(err).To(HaveOccurred())
		Expect(string(stderr)).Should(MatchRegexp("order of .* is smaller than required"))
	})
}
