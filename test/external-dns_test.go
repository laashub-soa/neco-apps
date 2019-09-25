package test

import (
	"encoding/json"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func testExternalDNS() {
	It("should be deployed successfully", func() {
		Eventually(func() error {
			stdout, _, err := ExecAt(boot0, "kubectl", "--namespace=external-dns",
				"get", "deployment/external-dns", "-o=json")
			if err != nil {
				return err
			}

			deployment := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, deployment)
			if err != nil {
				return err
			}

			if deployment.Status.AvailableReplicas != 1 {
				return fmt.Errorf("external-dns deployment's AvailableReplica is not 1: %d", int(deployment.Status.AvailableReplicas))
			}
			return nil
		}).Should(Succeed())
	})

	It("should create a DNS record", func() {
		By("deploying IngressRoute and its backend")
		ingressRouteBackend := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: testhttpd
  namespace: ingress
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: testhttpd
  template:
    metadata:
      labels:
        app.kubernetes.io/name: testhttpd
    spec:
      containers:
      - image: quay.io/cybozu/testhttpd:0
        name: testhttpd
      restartPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  name: testhttpd-external-dns
  namespace: ingress
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8000
  selector:
    app.kubernetes.io/name: testhttpd
---
apiVersion: crd.projectcalico.org/v1
kind: NetworkPolicy
metadata:
  name: ingress-httpdtest
  namespace: ingress
spec:
  order: 2000.0
  selector: app.kubernetes.io/name == 'testhttpd'
  types:
    - Ingress
  ingress:
    - action: Allow
      protocol: TCP
      destination:
        ports:
          - 8000
`
		_, stderr, err := ExecAtWithInput(boot0, []byte(ingressRouteBackend), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		By("waiting pods are ready")
		Eventually(func() error {
			stdout, _, err := ExecAt(boot0, "kubectl", "-n", "ingress", "get", "deployments/testhttpd", "-o", "json")
			if err != nil {
				return err
			}

			deployment := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, deployment)
			if err != nil {
				return err
			}

			if deployment.Status.ReadyReplicas != 1 {
				return errors.New("ReadyReplicas is not 1")
			}
			return nil
		}).Should(Succeed())

		domainName := testID + "-external-dns.gcp0.dev-ne.co"
		ingressRoute := fmt.Sprintf(`
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata:
  name: test
  namespace: ingress
spec:
  virtualhost:
    fqdn: %s
  routes:
    - match: /
      services:
        - name: testhttpd-external-dns
          port: 80
`, domainName)
		_, stderr, err = ExecAtWithInput(boot0, []byte(ingressRoute), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		By("resolving xxx.gcp0.dev-ne.co")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "nodes", "-o", "json")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			var nl corev1.NodeList
			err = json.Unmarshal(stdout, &nl)
			if err != nil {
				return err
			}
			if len(nl.Items) == 0 {
				return errors.New("nodes not found")
			}
			node := nl.Items[0]

			var nodeAddress string
			for _, addr := range node.Status.Addresses {
				if addr.Type != corev1.NodeInternalIP {
					continue
				}
				nodeAddress = addr.Address
			}
			if len(nodeAddress) == 0 {
				return errors.New("can not get NodeIP")
			}

			stdout, stderr, err = ExecAt(boot0, "ckecli", "ssh", nodeAddress, "dig", "+noall", "+answer", domainName)
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		By("cleaning up")
		_, stderr, err = ExecAtWithInput(boot0, []byte(ingressRoute), "kubectl", "delete", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)
		_, stderr, err = ExecAtWithInput(boot0, []byte(ingressRouteBackend), "kubectl", "delete", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)
	})
}
