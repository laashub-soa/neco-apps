package test

import (
	"encoding/json"
	"io/ioutil"

	argoappv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

// testSetup tests setup of Argo CD
func testSetup() {
	It("should install Argo CD", func() {
		execSafeAt(boot0, "ckecli", "kubernetes", "issue", ">", ".kube/config")
		execSafeAt(boot0, "kubectl", "create", "namespace", testID)

		data, err := ioutil.ReadFile("install.yaml")
		Expect(err).ShouldNot(HaveOccurred())
		stdout, stderr, err := execAtWithInput(boot0, data, "kubectl", "create", "-n", testID, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	})

	It("should install argocd CLI", func() {
		execSafeAt(boot0, "sudo",
			"env", "HTTP_PROXY=http://10.0.49.3:3128", "HTTPS_PROXY=http://10.0.49.3:3128",
			"rkt", "run",
			"--volume", "host,kind=host,source=/usr/local/bin",
			"--mount", "volume=host,target=/host",
			"quay.io/cybozu/argocd:0.11",
			"--user=0", "--group=0",
			"--exec", "/usr/local/argocd/install-tools")
	})

	It("should login to Argo CD", func() {
		By("getting password")
		// admin password is same as pod name
		var podList corev1.PodList
		Eventually(func() error {
			data := execSafeAt(boot0, "kubectl", "get", "pods", "-n", testID,
				"-l", "app=argocd-server", "-o", "json")
			return json.Unmarshal(data, &podList)
		}).Should(Succeed())
		Expect(podList.Items).ShouldNot(BeEmpty())
		password := podList.Items[0].Name

		By("getting node address")
		var nodeList corev1.NodeList
		data := execSafeAt(boot0, "kubectl", "get", "nodes", "-o", "json")
		err := json.Unmarshal(data, &nodeList)
		Expect(err).ShouldNot(HaveOccurred(), "data=%s", string(data))
		Expect(nodeList.Items).ShouldNot(BeEmpty())
		node := nodeList.Items[0]

		var nodeAddress string
		for _, addr := range node.Status.Addresses {
			if addr.Type != corev1.NodeInternalIP {
				continue
			}
			nodeAddress = addr.Address
		}
		Expect(nodeAddress).ShouldNot(BeNil())

		execSafeAt(boot0, "argocd", "login", nodeAddress+":30080",
			"--username", "admin", "--password", password)
	})

	It("should setup application", func() {
		By("creating guestbook")
		execSafeAt(boot0, "argocd", "app", "create", "guestbook",
			"--repo", "https://github.com/argoproj/argocd-example-apps",
			"--path", "kustomize-guestbook", "--dest-server", "https://kubernetes.default.svc",
			"--dest-namespace", testID)
		execSafeAt(boot0, "argocd", "app", "sync", "guestbook")

		By("checking guestbook status")
		data := execSafeAt(boot0, "argocd", "app", "get", "guestbook", "-o", "json")
		var app argoappv1.Application
		err := json.Unmarshal(data, &app)
		Expect(err).ShouldNot(HaveOccurred(), "data=%s", string(data))

		for _, r := range app.Status.Resources {
			Expect(r.Status).Should(Equal(argoappv1.SyncStatusCodeSynced))
			Expect(r.Health.Status).Should(Equal(argoappv1.HealthStatusHealthy))
		}
	})
}
