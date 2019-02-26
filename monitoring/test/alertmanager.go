package monitoring

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cybozu-go/neco-ops/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const alertmanagerSecret = `
route:
  receiver: slack
  group_wait: 5s # Send a notification after 5 seconds
  routes:
  - receiver: slack
    continue: true # Continue notification to next receiver

# Receiver configurations
receivers:
- name: slack
  slack_configs:
  - channel: '#test'
    api_url: https://hooks.slack.com/services/XXX/XXX
    icon_url: https://avatars3.githubusercontent.com/u/3380462 # Prometheus icon
    http_config:
      proxy_url: http://squid.internet-egress.svc.cluster.local:3128
`

func testAlertmanager() {
	It("should be deployed successfully", func() {
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=monitoring",
				"get", "deployment/alertmanager", "-o=json")
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

	It("should reply successfully", func() {
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=monitoring",
				"get", "pods", "--selector=app=alertmanager", "-o=json")
			if err != nil {
				return err
			}
			podList := new(corev1.PodList)
			err = json.Unmarshal(stdout, podList)
			if err != nil {
				return err
			}
			if len(podList.Items) != 1 {
				return errors.New("alertmanager pod doesn't exist")
			}
			podName := podList.Items[0].Name

			_, _, err = test.ExecAt(test.Boot0, "kubectl", "--namespace=monitoring", "exec",
				podName, "curl", "http://localhost:9093/-/healthy")
			if err != nil {
				return err
			}
			return nil
		}).Should(Succeed())
	})
}
