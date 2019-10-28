package test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

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
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		By("requesting to web UI")
		resp, err := client.Get("https://argocd.gcp0.dev-ne.co")
		Expect(err).ShouldNot(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		contentType := resp.Header.Get("Content-type")
		Expect(contentType).NotTo(BeEmpty())
		Expect(contentType).NotTo(Equal("application/grpc"))
		fmt.Printf("Status: %d, Content-type: %s\n", resp.StatusCode, contentType)

		By("requesting to dex server via argocd-server")
		resp, err = client.Get("https://argocd.gcp0.dev-ne.co/api/dex")
		Expect(err).ShouldNot(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		contentType = resp.Header.Get("Content-type")
		Expect(contentType).NotTo(BeEmpty())
		Expect(contentType).NotTo(Equal("application/grpc"))
		fmt.Printf("Status: %d, Content-type: %s\n", resp.StatusCode, contentType)
	})
}
