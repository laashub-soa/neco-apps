package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func testArgoCDIngress() {
	It("should login via IngressRoute", func() {
		By("getting the ip address of the contour LoadBalancer")
		stdout, _, err := ExecAt(boot0, "kubectl", "--namespace=ingress", "get", "service/contour-bastion", "-o=json")
		Expect(err).ShouldNot(HaveOccurred())

		svc := new(corev1.Service)
		err = json.Unmarshal(stdout, svc)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(svc.Status.LoadBalancer.Ingress)).To(Equal(1))
		lbIP := svc.Status.LoadBalancer.Ingress[0].IP

		By("adding loadbalancer address entry to /etc/hosts")
		_, stderr, err := ExecAt(boot0, "sudo", "bash", "-c", "'echo "+lbIP+" argocd.gcp0.dev-ne.co >> /etc/hosts'")
		Expect(err).ShouldNot(HaveOccurred(), "stderr: %s", stderr)

		By("logging in to Argo CD")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "argocd", "login", "argocd.gcp0.dev-ne.co",
				"--insecure", "--username", "admin", "--password", loadArgoCDPassword())
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())
	})

	It("should communicate with http protocol", func() {
		By("requesting to web UI")
		stdout, stderr, err := ExecAt(boot0,
			"curl", "-skI", "https://argocd.gcp0.dev-ne.co",
			"-o", "/dev/nulll",
			"-w", `'%{http_code}\n%{content_type}'`,
		)
		Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
		s := strings.Split(string(stdout), "\n")
		Expect(s[0]).To(Equal(strconv.Itoa(http.StatusOK)))
		Expect(s[1]).To(Equal("application/grpc"))

		By("requesting to dex server via argocd-server")
		stdout, stderr, err = ExecAt(boot0,
			"curl", "-skI", "https://argocd.gcp0.dev-ne.co/api/dex",
			"-o", "/dev/null",
			"-w", `'%{http_code}\n%{content_type}'`,
		)
		Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
		s = strings.Split(string(stdout), "\n")
		Expect(s[0]).To(Equal(strconv.Itoa(http.StatusOK)))
		Expect(s[1]).To(Equal("application/grpc"))
	})
}
