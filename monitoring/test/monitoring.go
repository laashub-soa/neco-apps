package monitoring

import (
	"encoding/json"
	"fmt"

	argoappv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/cybozu-go/neco-ops/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testMonitoring() {
	It("should be deployed by Argo CD", func() {
		By("setting application parameters and synchronizing")
		stdout, stderr, err := test.ExecAtWithInput(test.Boot0, []byte(alertmanagerSecret), "dd", "of=alertmanager.yaml")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
		stdout, stderr, err = test.ExecAt(test.Boot0, "kubectl", "create", "namespace", "monitoring")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
		stdout, stderr, err = test.ExecAt(test.Boot0, "kubectl", "--namespace=monitoring", "create", "secret",
			"generic", "alertmanager", "--from-file", "alertmanager.yaml")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
		Eventually(func() error {
			// in some case, doing 'argocd app set' once is not sufficient somehow...
			stdout, stderr, err := test.ExecAt(test.Boot0, "argocd", "app", "set", "monitoring", "--revision", test.CommitID)
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			stdout, stderr, err = test.ExecAt(test.Boot0, "argocd", "app", "sync", "monitoring", "--timeout", "60")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		By("checking monitoring status")
		Eventually(func() error {
			stdout, stderr, err := test.ExecAt(test.Boot0,
				"kubectl", "get", "app", "monitoring", "-n", test.ArgoCDNamespace, "-o", "json")
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
