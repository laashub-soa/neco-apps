package monitoring

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cybozu-go/neco-ops/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func testPrometheus() {
	It("should be deployed successfully", func() {
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=monitoring",
				"get", "deployment/prometheus", "-o=json")
			if err != nil {
				return err
			}
			deployment := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, deployment)
			if err != nil {
				return err
			}

			if int(deployment.Status.AvailableReplicas) != 1 {
				return fmt.Errorf("AvailableReplicas is not 1: %d", int(deployment.Status.AvailableReplicas))
			}
			return nil
		}).Should(Succeed())
	})
	var podName string
	It("should reply successfully", func() {
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=monitoring",
				"get", "pods", "--selector=app=prometheus", "-o=json")
			if err != nil {
				return err
			}
			podList := new(corev1.PodList)
			err = json.Unmarshal(stdout, podList)
			if err != nil {
				return err
			}
			if len(podList.Items) != 1 {
				return errors.New("prometheus pod doesn't exist")
			}
			podName = podList.Items[0].Name

			_, _, err = test.ExecAt(test.Boot0, "kubectl", "--namespace=monitoring", "exec",
				podName, "curl", "http://localhost:9090/api/v1/alerts")
			if err != nil {
				return err
			}
			return nil
		}).Should(Succeed())
	})

	It("should find endpoint", func() {
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=monitoring", "exec",
				podName, "curl", "http://localhost:9090/api/v1/targets")
			if err != nil {
				return err
			}

			var response struct {
				TargetsResult promv1.TargetsResult `json:"data"`
			}
			err = json.Unmarshal(stdout, &response)
			if err != nil {
				return err
			}

			for _, target := range response.TargetsResult.Active {
				if value, ok := target.Labels["kubernetes_name"]; ok {
					if value == "prometheus-node-targets" {
						return nil
					}
				}
			}
			return errors.New("cannot found node targets")
		}).Should(Succeed())
	})
}
