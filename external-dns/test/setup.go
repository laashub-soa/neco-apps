package test

import (
	"encoding/json"
	"fmt"

	argoappv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/cybozu-go/neco-ops/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testSetup() {
	It("should deploy external-dns by Argo CD", func() {
		By("registering secrets")
		_, stderr, err := test.ExecAt(test.Boot0, "kubectl", "create", "namespace", "external-dns")
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		_, stderr, err = test.ExecAt(test.Boot0, "kubectl", "--namespace=external-dns", "create", "secret",
			"generic", "external-dns", "--from-file=account.json")
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)

		By("synchronizing external-dns")
		Eventually(func() error {
			stdout, stderr, err := test.ExecAt(test.Boot0, "argocd", "app", "set", "external-dns", "--revision", test.CommitID)
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			stdout, stderr, err = test.ExecAt(test.Boot0, "argocd", "app", "sync", "external-dns")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		By("checking external-dns status")
		Eventually(func() error {
			stdout, stderr, err := test.ExecAt(test.Boot0,
				"kubectl", "get", "app", "external-dns", "-n", test.ArgoCDNamespace, "-o", "json")
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
