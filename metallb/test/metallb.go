package metallb

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	argoappv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/cybozu-go/neco-ops/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func testMetalLB() {
	It("should deploy metallb by Argo CD", func() {
		By("synchronizing metallb")
		Eventually(func() error {
			stdout, stderr, err := test.ExecAt(test.Boot0, "argocd", "app", "set", "metallb", "--revision", test.CommitID)
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			stdout, stderr, err = test.ExecAt(test.Boot0, "argocd", "app", "sync", "metallb", "--timeout", "20")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		By("checking metallb status")
		Eventually(func() error {
			stdout, stderr, err := test.ExecAt(test.Boot0,
				"kubectl", "get", "app", "metallb", "-n", test.ArgoCDNamespace, "-o", "json")
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
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=metallb-system",
				"get", "daemonsets/speaker", "-o=json")
			if err != nil {
				return err
			}
			ds := new(appsv1.DaemonSet)
			err = json.Unmarshal(stdout, ds)
			if err != nil {
				return err
			}

			if ds.Status.DesiredNumberScheduled <= 0 {
				return errors.New("speaker daemonset's desiredNumberScheduled is not updated")
			}

			if ds.Status.DesiredNumberScheduled != ds.Status.NumberAvailable {
				return fmt.Errorf("not all nodes running speaker daemonset: %d", ds.Status.NumberAvailable)
			}
			return nil
		}).Should(Succeed())

		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "--namespace=metallb-system",
				"get", "deployments/controller", "-o=json")
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

	It("should deploy load balancer type service", func() {
		By("deployment Pods")
		_, stderr, err := test.ExecAt(test.Boot0, "kubectl", "run", "nginx", "--replicas=2", "--image=nginx")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		By("waiting pods are ready")
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "get", "deployments/nginx", "-o", "json")
			if err != nil {
				return err
			}

			deployment := new(appsv1.Deployment)
			err = json.Unmarshal(stdout, deployment)
			if err != nil {
				return err
			}

			if deployment.Status.ReadyReplicas != 2 {
				return errors.New("ReadyReplicas is not 2")
			}
			return nil
		}).Should(Succeed())

		By("create Service")
		targetIP := "10.72.32.29"
		loadBalancer := `
kind: Service
apiVersion: v1
metadata:
  name: nginx
  namespace: default
spec:
  selector:
    run: nginx
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: LoadBalancer
  loadBalancerIP: 10.72.32.29
  externalTrafficPolicy: Local
`
		_, stderr, err = test.ExecAtWithInput(test.Boot0, []byte(loadBalancer), "kubectl", "create", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		By("waiting service are ready")
		Eventually(func() error {
			stdout, _, err := test.ExecAt(test.Boot0, "kubectl", "get", "service/nginx", "-o", "json")
			if err != nil {
				return err
			}

			service := new(corev1.Service)
			err = json.Unmarshal(stdout, service)
			if err != nil {
				return err
			}

			if len(service.Status.LoadBalancer.Ingress) == 0 {
				return errors.New("LoadBalancer status is not updated")
			}

			actualIP := service.Status.LoadBalancer.Ingress[0].IP
			if actualIP != targetIP {
				return fmt.Errorf("LoadBalancer is not %s, %s", targetIP, actualIP)
			}
			return nil
		}).Should(Succeed())

		By("access service from boot-0")
		Eventually(func() error {
			_, _, err := test.ExecAt(test.Boot0, "curl", targetIP, "-m", "5")
			return err
		}).Should(Succeed())

		By("access service from external")
		Eventually(func() error {
			cmd := exec.Command("sudo", "nsenter", "-n", "-t", test.ExternalPid, "curl", targetIP, "-m", "5")
			return cmd.Run()
		}).Should(Succeed())
	})
}
