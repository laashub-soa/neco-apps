package test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	contourv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
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

	It("should default/validate Contour HTTPProxy", func() {
		httpProxyYAML := `
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: bad
  namespace: default
spec:
  virtualhost:
    fqdn: bad.test-admission.gcp0.dev-ne.co
  routes:
    - conditions:
        - prefix: /
      services:
        - name: dummy
          port: 80
`
		By("creating HTTPProxy without annotations")
		stdout, stderr, err := ExecAtWithInput(boot0, []byte(httpProxyYAML), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)

		stdout, stderr, err = ExecAt(boot0, "kubectl", "get", "-n", "default", "httpproxy/bad", "-o", "json")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)

		hp := new(contourv1.HTTPProxy)
		err = json.Unmarshal(stdout, hp)
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, err: %v", stdout, err)
		Expect(hp.Annotations).To(HaveKeyWithValue("kubernetes.io/ingress.class", "forest"))

		By("updating HTTPProxy to remove annotations")
		stdout, stderr, err = ExecAt(boot0, "kubectl", "annotate", "-n", "default", "httpproxy/bad", "kubernetes.io/ingress.class-")
		Expect(err).To(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

		stdout, stderr, err = ExecAtWithInput(boot0, []byte(httpProxyYAML), "kubectl", "delete", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	})
}
