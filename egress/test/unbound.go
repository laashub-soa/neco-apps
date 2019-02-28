package egress

import (
	"encoding/json"
	"fmt"

	"github.com/cybozu-go/neco-ops/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
)

func testUnbound() {
	It("should be deployed successfully", func() {
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=internet-egress",
				"get", "deployments/unbound", "-o=json")
			if err != nil {
				return err
			}

			deployment := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, deployment)
			if err != nil {
				return err
			}

			if int(deployment.Status.AvailableReplicas) != 2 {
				return fmt.Errorf("AvailableReplicas is not 2: %d", int(deployment.Status.AvailableReplicas))
			}
			return nil
		}).Should(Succeed())
	})

	It("should resolve www.cybozu.com", func() {
		By("running a test pod")
		podName := "unbound-test"
		test.ExecSafeAt(test.Boot0, "kubectl", "run", podName,
			"--image=$(ckecli images | grep quay.io/cybozu/cke-tools)",
			"--generator=run-pod/v1", "--", "/bin/sleep", "infinity")

		By("executing getent hosts www.cybozu.com")
		Eventually(func() error {
			_, _, err := test.ExecAt(test.Boot0, "kubectl", "exec", podName,
				"getent", "hosts", "www.cybozu.com")
			return err
		}).Should(Succeed())

		test.ExecSafeAt(test.Boot0, "kubectl", "delete", "pod", podName)
	})
}
