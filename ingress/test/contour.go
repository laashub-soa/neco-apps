package ingress

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/cybozu-go/neco-ops/test"
	"github.com/kubernetes-incubator/external-dns/endpoint"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func testContour() {
	It("should be deployed successfully", func() {
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=ingress",
				"get", "deployment/contour", "-o=json")
			if err != nil {
				return err
			}

			deployment := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, deployment)
			if err != nil {
				return err
			}

			if deployment.Status.AvailableReplicas != 2 {
				return fmt.Errorf("contour deployment's AvailableReplica is not 2: %d", int(deployment.Status.AvailableReplicas))
			}
			return nil
		}).Should(Succeed())
	})

	It("should deploy IngressRoute", func() {
		By("deployment Pods")
		_, stderr, err := test.ExecAt(test.Boot0, "kubectl", "-n", "test-ingress", "run", "testhttpd", "--image=quay.io/cybozu/testhttpd:0", "--replicas=2")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		By("waiting pods are ready")
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "-n", "test-ingress", "get", "deployments/testhttpd", "-o", "json")
			if err != nil {
				return err
			}

			deployment := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, deployment)
			if err != nil {
				return err
			}

			if deployment.Status.ReadyReplicas != 2 {
				return errors.New("ReadyReplicas is not 2")
			}
			return nil
		}).Should(Succeed())

		_, stderr, err = test.ExecAt(test.Boot0, "kubectl", "-n", "test-ingress", "expose", "deployment", "testhttpd", "--port=80", "--target-port=8000", "--name=testhttpd")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		By("creating IngressRoute")
		fqdn := "test-ingress.gcp0.dev-ne.co"
		ingressRoute := fmt.Sprintf(`
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata:
  name: root
  namespace: test-ingress
spec:
  virtualhost:
    fqdn: %s
  routes:
    - match: /testhttpd
      services:
        - name: testhttpd
          port: 80
`, fqdn)
		_, stderr, err = test.ExecAtWithInput(test.Boot0, []byte(ingressRoute), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		By("getting contour service")
		var targetIP string
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "get", "-n", "ingress", "service/contour-global", "-o", "json")
			if err != nil {
				return err
			}

			service := new(corev1.Service)
			err = json.Unmarshal(stdout, service)
			if err != nil {
				return err
			}

			if len(service.Status.LoadBalancer.Ingress) < 1 {
				return errors.New("LoadBalancerIP is not assigned")
			}
			targetIP = service.Status.LoadBalancer.Ingress[0].IP
			if len(targetIP) == 0 {
				return errors.New("LoadBalancerIP is empty")
			}
			return nil
		}).Should(Succeed())

		By("accessing service from operation")
		Eventually(func() error {
			cmd := exec.Command("curl", "--header", "Host: "+fqdn, targetIP+"/testhttpd", "-m", "5", "--fail")
			return cmd.Run()
		}).Should(Succeed())

		By("generating DNSEndpoint automatically")
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "get", "-n", "ingress", "dnsendpoint/cp-root", "-o", "json")
			if err != nil {
				return err
			}
			de := new(endpoint.DNSEndpoint)
			err = json.Unmarshal(stdout, de)
			if err != nil {
				return err
			}
			if len(de.Spec.Endpoints) == 0 {
				return errors.New("len(de.Spec.Endpoints) == 0")
			}

			expectedIP := de.Spec.Endpoints[0].Targets[0]
			stdout, _, err = test.ExecAt(test.Boot0, "kubectl", "get", "-n=ingress", "svc/contour-forest", "-o=template",
				`--template="{{(index .status.loadBalancer.ingress 0).ip}}"`)
			if err != nil {
				return err
			}
			actualIP := string(stdout)
			if expectedIP != actualIP {
				return fmt.Errorf("expected IP is (%s), but actual is (%s)", expectedIP, actualIP)
			}
			return nil
		}).Should(Succeed())
	})
}
