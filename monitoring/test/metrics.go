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
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

func testMetrics() {
	It("should be up all scraping", func() {
		var podName string
		By("retrieving prometheus podName")
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
		stdout, stderr, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=monitoring",
			"get", "configmap", "-o=json")
		Expect(err).NotTo(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		cmList := new(corev1.ConfigMapList)
		err = json.Unmarshal(stdout, cmList)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(cmList.Items)).To(Equal(1))
		data, ok := cmList.Items[0].Data["prometheus.yaml"]
		Expect(ok).NotTo(BeFalse())
		promConfig := new(promconfig.Config)
		err = yaml.Unmarshal([]byte(data), promConfig)
		Expect(err).NotTo(HaveOccurred())
		var jobNames []model.LabelName
		for _, sc := range promConfig.ScrapeConfigs {
			jobNames = append(jobNames, model.LabelName(sc.JobName))
		}

		By("checking discovered active labels and statuses")
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

			for _, jobName := range jobNames {
				for _, target := range response.TargetsResult.Active {
					if _, ok := target.Labels[jobName]; ok {
						if target.Health != promv1.HealthGood {
							return fmt.Errorf("target is not up, job_name: %s", jobName)
						}
					}
				}
			}
			return nil
		}).Should(Succeed())
	})
}
