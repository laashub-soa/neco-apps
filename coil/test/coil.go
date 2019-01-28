package coil

import (
	"encoding/json"
	"errors"
	"fmt"

	argoappv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/cybozu-go/neco-ops/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func testCoil() {
	It("should deploy coil by Argo CD", func() {
		By("synchronizing coil")
		Eventually(func() error {
			// To apply target commitID for auto-sync enabled app, kubectl patch allows app to change targetRevision.
			stdout, stderr, err := test.ExecAt(test.Boot0, "kubectl", "patch",
				"-n", test.ArgoCDNamespace, "app", "coil", "--type=merge",
				"-p", `'{"spec":{"source":{"targetRevision":"`+test.CommitID+`"}}}'`)
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			stdout, stderr, err = test.ExecAt(test.Boot0, "argocd", "app", "sync", "coil", "--timeout", "20")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		By("checking coil status")
		Eventually(func() error {
			stdout, stderr, err := test.ExecAt(test.Boot0,
				"kubectl", "get", "app", "coil", "-n", test.ArgoCDNamespace, "-o", "json")
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

	It("should be deployed successfully", func() {
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=kube-system",
				"get", "daemonsets/coil-node", "-o=json")
			if err != nil {
				return err
			}

			daemonset := new(appsv1.DaemonSet)
			err = json.Unmarshal(stdout, daemonset)
			if err != nil {
				return err
			}

			if int(daemonset.Status.NumberReady) != 5 {
				return errors.New("NumberReady is not 5")
			}
			return nil
		}).Should(Succeed())

		By("checking exist IP address pool")
		stdout, stderr, err := test.ExecAt(test.Boot0,
			"kubectl", "--namespace=kube-system", "get", "pods", "--selector=k8s-app=coil-controllers", "-o=json")
		Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		podList := new(corev1.PodList)
		err = json.Unmarshal(stdout, podList)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(1))
		podName := podList.Items[0].Name

		stdout, stderr, err = test.ExecAt(test.Boot0,
			"kubectl", "--namespace=kube-system", "exec", podName, "/coilctl", "pool", "list")
		Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	})
}
