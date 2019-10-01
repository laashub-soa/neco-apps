package test

import (
	"encoding/json"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func testGatekeeper() {
	It("should make gatekeeper ready", func() {
		By("confirming that controller manager pod is ready")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "-n=gatekeeper-system", "pods", "gatekeeper-controller-manager-0", "-o", "json")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			var pod corev1.Pod
			err = json.Unmarshal(stdout, &pod)
			if err != nil {
				return err
			}

			if len(pod.Status.Conditions) == 0 {
				return errors.New("conditions not found")
			}

			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
					return nil
				}
			}
			return errors.New("condition 'Ready' is not found")
		}).Should(Succeed())

		By("confirming that constraint templates is established")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "crd", "constrainttemplates.templates.gatekeeper.sh", "-o", "json")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			var crd struct {
				Status struct {
					Conditions []struct {
						Type   string `json:"type"`
						Status string `json:"status"`
					} `json:"conditions"`
				} `json:"status"`
			}
			err = json.Unmarshal(stdout, &crd)
			if err != nil {
				return err
			}

			if len(crd.Status.Conditions) == 0 {
				return errors.New("conditions not found")
			}

			for _, cond := range crd.Status.Conditions {
				if cond.Type == "Established" && cond.Status == "True" {
					return nil
				}
			}
			return errors.New("condition 'Established' is not found")
		}).Should(Succeed())

		By("confirming validating webhook can be got")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(
				boot0, "kubectl", "get",
				"validatingwebhookconfigurations.admissionregistration.k8s.io", "validation.gatekeeper.sh",
			)
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

	})

	It("should validate Service resource according to policy", func() {
		By("registering constraint template")
		constraintTemplete := `
apiVersion: templates.gatekeeper.sh/v1beta1
kind: ConstraintTemplate
metadata:
  name: k8srequiredlabels
spec:
  crd:
    spec:
      names:
        kind: K8sRequiredLabels
      validation:
        openAPIV3Schema:
          properties:
            labels:
              type: array
              items:
                type: string
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |
        package k8srequiredlabels

        violation[{"msg": msg, "details": {"missing_labels": missing}}] {
          provided := {label | input.review.object.metadata.labels[label]}
          required := {label | label := input.parameters.labels[_]}
          missing := required - provided
          count(missing) > 0
          msg := sprintf("you must provide labels: %v", [missing])
        }
`
		stdout, stderr, err := ExecAtWithInput(boot0, []byte(constraintTemplete), "kubectl", "apply", "-f", "-")
		Expect(err).To(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

		By("confirming that required labels constraint is established")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "crd", "k8srequiredlabels.constraints.gatekeeper.sh", "-o", "json")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			var crd struct {
				Status struct {
					Conditions []struct {
						Type   string `json:"type"`
						Status string `json:"status"`
					} `json:"conditions"`
				} `json:"status"`
			}
			err = json.Unmarshal(stdout, &crd)
			if err != nil {
				return err
			}

			if len(crd.Status.Conditions) == 0 {
				return errors.New("conditions not found")
			}

			for _, cond := range crd.Status.Conditions {
				if cond.Type == "Established" && cond.Status == "True" {
					return nil
				}
			}
			return errors.New("condition 'Established' is not found")
		}).Should(Succeed())

		By("registering constraint")
		constraint := `
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sRequiredLabels
metadata:
  name: ns-must-have-gk
spec:
  match:
    kinds:
      - apiGroups: [""]
        kinds: ["Namespace"]
  parameters:
    labels: ["gatekeeper"]
`
		stdout, stderr, err = ExecAtWithInput(boot0, []byte(constraint), "kubectl", "apply", "-f", "-")
		Expect(err).To(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

		By("waiting for constrains enforced")
		Eventually(func() error {
			stdout, stderr, err := ExecAtWithInput(boot0, []byte(constraint), "kubectl", "get", "k8srequiredlabels", "ns-must-have-gk", "-o", "json")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			var constraint struct {
				Status struct {
					ByPod []struct {
						Enforced bool `json:"enforced"`
					} `json:"byPod"`
				} `json:"status"`
			}

			err = json.Unmarshal(stdout, &constraint)
			if err != nil {
				return fmt.Errorf("data: %s, err: %v", string(stdout), err)
			}
			if len(constraint.Status.ByPod) < 1 {
				return errors.New("byPod is empty")
			}
			if constraint.Status.ByPod[0].Enforced != true {
				return errors.New("not enforced")
			}
			return nil
		}).Should(Succeed())

		By("creating bad Namespace")
		badNSYAML := `
apiVersion: v1
kind: Namespace
metadata:
  name: bad
`
		stdout, stderr, err = ExecAtWithInput(boot0, []byte(badNSYAML), "kubectl", "apply", "-f", "-")
		Expect(err).To(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

		By("creating good Namespace")
		goodNSYAML := `
apiVersion: v1
kind: Namespace
metadata:
  name: good
  labels:
    gatekeeper: defined
`
		stdout, stderr, err = ExecAtWithInput(boot0, []byte(goodNSYAML), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

		By("unregistering policy")
		stdout, stderr, err = ExecAt(boot0, "kubectl", "delete", "K8sRequiredLabels", "ns-must-have-gk")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

		By("creating formerly-bad Namespace")
		stdout, stderr, err = ExecAtWithInput(boot0, []byte(badNSYAML), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

		By("deleting Namespaces")
		stdout, stderr, err = ExecAt(boot0, "kubectl", "delete", "namespace", "bad")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
		stdout, stderr, err = ExecAt(boot0, "kubectl", "delete", "namespace", "good")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
	})
}
