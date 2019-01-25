package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	argoappv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

// testSetup tests setup of Argo CD
func testSetup() {
	It("should be ready K8s cluster after loading snapshot", func() {
		By("re-issuing kubeconfig")
		Eventually(func() error {
			_, _, err := execAt(boot0, "ckecli", "kubernetes", "issue", ">", ".kube/config")
			if err != nil {
				return err
			}
			return nil
		}).Should(Succeed())

		By("waiting nodes")
		Eventually(func() error {
			stdout, _, err := execAt(boot0, "kubectl", "get", "nodes", "-o", "json")
			if err != nil {
				return err
			}

			var nl corev1.NodeList
			err = json.Unmarshal(stdout, &nl)
			if err != nil {
				return err
			}

			if len(nl.Items) != 5 {
				return fmt.Errorf("too few nodes: %d", len(nl.Items))
			}

		OUTER:
			for _, n := range nl.Items {
				for _, cond := range n.Status.Conditions {
					if cond.Type != corev1.NodeReady {
						continue
					}
					if cond.Status != corev1.ConditionTrue {
						return fmt.Errorf("node %s is not ready", n.Name)
					}
					continue OUTER
				}

				return fmt.Errorf("node %s has no readiness status", n.Name)
			}

			return nil
		}).Should(Succeed())
	})

	It("should install Argo CD", func() {
		execSafeAt(boot0, "kubectl", "create", "namespace", argoCDNamespace)

		data, err := ioutil.ReadFile("install.yaml")
		Expect(err).ShouldNot(HaveOccurred())
		stdout, stderr, err := execAtWithInput(boot0, data, "kubectl", "apply", "-n", argoCDNamespace, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	})

	It("should install argocd CLI", func() {
		execSafeAt(boot0, "sudo",
			"env", "HTTP_PROXY=http://10.0.49.3:3128", "HTTPS_PROXY=http://10.0.49.3:3128",
			"rkt", "run",
			"--insecure-options=image",
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
			data := execSafeAt(boot0, "kubectl", "get", "pods", "-n", argoCDNamespace,
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

		By("getting node port")
		var svc corev1.Service
		data = execSafeAt(boot0, "kubectl", "get", "svc/argocd-server", "-n", argoCDNamespace, "-o", "json")
		err = json.Unmarshal(data, &svc)
		Expect(err).ShouldNot(HaveOccurred(), "data=%s", string(data))
		Expect(svc.Spec.Ports).ShouldNot(BeEmpty())

		var nodePort string
		for _, port := range svc.Spec.Ports {
			if port.Name != "http" {
				continue
			}
			nodePort = strconv.Itoa(int(port.NodePort))
		}
		Expect(nodePort).ShouldNot(BeNil())

		By("logging to Argo CD")
		Eventually(func() error {
			stdout, stderr, err := execAt(boot0, "argocd", "login", nodeAddress+":"+nodePort,
				"--insecure", "--username", "admin", "--password", password)
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())
	})

	It("should setup application", func() {
		By("creating guestbook")
		execSafeAt(boot0, "kubectl", "create", "namespace", testID)
		execSafeAt(boot0, "argocd", "app", "create", "guestbook",
			"--repo", "https://github.com/argoproj/argocd-example-apps",
			"--path", "kustomize-guestbook", "--dest-server", "https://kubernetes.default.svc",
			"--dest-namespace", testID)
		execSafeAt(boot0, "argocd", "app", "sync", "guestbook")

		By("checking guestbook status")
		Eventually(func() error {
			data := execSafeAt(boot0, "argocd", "app", "get", "guestbook", "-o", "json")
			var app argoappv1.Application
			err := json.Unmarshal(data, &app)
			if err != nil {
				return err
			}

			for _, r := range app.Status.Resources {
				if r.Status != argoappv1.SyncStatusCodeSynced {
					return fmt.Errorf("app is not yet Synced: %s", r.Status)
				}
				if r.Health.Status != argoappv1.HealthStatusHealthy {
					return fmt.Errorf("app is not yet Healthy: %s", r.Health.Status)
				}
			}
			return nil
		}).Should(Succeed())
	})
}
