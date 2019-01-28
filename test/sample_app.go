package test

import (
	"encoding/json"
	"fmt"

	argoappv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// testSampleApp tests sample app deployment by Argo CD
func testSampleApp() {
	It("should deploy guestbook sample app by Argo CD", func() {
		By("synchronizing guestbook sample app")
		Eventually(func() error {
			// To apply target commitID for auto-sync enabled app, kubectl patch allows app to change targetRevision.
			stdout, stderr, err := ExecAt(Boot0, "kubectl", "patch",
				"-n", ArgoCDNamespace, "app", "guestbook", "--type=merge",
				"-p", `'{"spec":{"source":{"targetRevision":"`+CommitID+`"}}}'`)
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			stdout, stderr, err = ExecAt(Boot0, "argocd", "app", "sync", "guestbook", "--timeout", "20")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		By("checking guestbook sample app status")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(Boot0, "kubectl", "get", "app", "guestbook", "-n", ArgoCDNamespace, "-o", "json")
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
