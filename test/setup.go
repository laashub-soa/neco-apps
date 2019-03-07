package test

import (
	"encoding/json"
	"errors"
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
			_, _, err := ExecAt(Boot0, "ckecli", "kubernetes", "issue", ">", ".kube/config")
			if err != nil {
				return err
			}
			return nil
		}).Should(Succeed())

		By("waiting nodes")
		Eventually(func() error {
			stdout, _, err := ExecAt(Boot0, "kubectl", "get", "nodes", "-o", "json")
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
		data, err := ioutil.ReadFile("install.yaml")
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(func() error {
			stdout, stderr, err := ExecAtWithInput(Boot0, data, "kubectl", "apply", "-n", ArgoCDNamespace, "-f", "-")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())
	})

	It("should login to Argo CD", func() {
		By("getting password")
		// admin password is same as pod name
		var podList corev1.PodList
		Eventually(func() error {
			data := ExecSafeAt(Boot0, "kubectl", "get", "pods", "-n", ArgoCDNamespace,
				"-l", "app=argocd-server", "-o", "json")
			err := json.Unmarshal(data, &podList)
			if err != nil {
				return err
			}
			if podList.Items == nil {
				return errors.New("podList.Items is nil")
			}
			if len(podList.Items) != 1 {
				return fmt.Errorf("podList.Items is not 1: %d", len(podList.Items))
			}
			return nil
		}).Should(Succeed())

		password := podList.Items[0].Name

		By("getting node address")
		var nodeList corev1.NodeList
		data := ExecSafeAt(Boot0, "kubectl", "get", "nodes", "-o", "json")
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
		data = ExecSafeAt(Boot0, "kubectl", "get", "svc/argocd-server", "-n", ArgoCDNamespace, "-o", "json")
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

		By("logging in to Argo CD")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(Boot0, "argocd", "login", nodeAddress+":"+nodePort,
				"--insecure", "--username", "admin", "--password", password)
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())
	})

	It("should setup Argo CD application as Argo CD app", func() {
		By("creating Argo CD app")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(Boot0, "argocd", "app", "create", "argocd-config",
				"--repo", "https://github.com/cybozu-go/neco-ops.git",
				"--path", "argocd-config/overlays/gcp",
				"--dest-namespace", ArgoCDNamespace,
				"--dest-server", "https://kubernetes.default.svc",
				"--revision", CommitID)
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		By("synchronizing Argo CD app")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(Boot0, "argocd", "app", "sync", "argocd-config", "--timeout", "20")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		By("checking argocd-config app status")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(Boot0, "kubectl", "get", "app", "argocd-config", "-n", ArgoCDNamespace, "-o", "json")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			var app argoappv1.Application
			err = json.Unmarshal(stdout, &app)
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
