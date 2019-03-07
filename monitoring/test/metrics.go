package monitoring

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cybozu-go/neco-ops/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	promconfig "github.com/prometheus/prometheus/config"
	corev1 "k8s.io/api/core/v1"
)

func testMetrics() {
	It("should be up all scraping", func() {
		var podName string
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
			return nil
		}).Should(Succeed())

		By("retrieving job_name from prometheus.yaml")
		promConfig := new(promconfig.Config)
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=monitoring",
				"get", "configmap", "--selector=app=prometheus", "-o=json")
			if err != nil {
				return err
			}

			cmList := new(corev1.ConfigMapList)
			err = json.Unmarshal(stdout, cmList)
			if err != nil {
				return err
			}
			if len(cmList.Items) != 1 {
				return fmt.Errorf("configMap is not 1, %d", len(cmList.Items))
			}

			data, ok := cmList.Items[0].Data["prometheus.yaml"]
			if !ok {
				return errors.New("prometheus.yaml does not exist")

			}
			err = json.Unmarshal([]byte(data), promConfig)
			if err != nil {
				return err
			}
			return nil
		}).Should(Succeed())

		var jobNames []model.LabelName
		for _, sc := range promConfig.ScrapeConfigs {
			jobNames = append(jobNames, model.LabelName(sc.JobName))
		}

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
				for _, jobName := range jobNames {
					if _, ok := target.Labels[jobName]; ok {
						if target.Health == promv1.HealthGood {
							return nil
						}
					}
				}
			}
			return errors.New("metrics is not health")
		}).Should(Succeed())
	})
}
