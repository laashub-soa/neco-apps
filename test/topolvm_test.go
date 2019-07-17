package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func testTopoLVM() {
	ns := "test-topolvm"
	It("should create test-topolvm namespace", func() {
		ExecSafeAt(boot0, "kubectl", "delete", "namespace", ns, "--ignore-not-found=true")
		ExecSafeAt(boot0, "kubectl", "create", "namespace", ns)
	})

	It("should be mounted in specified path", func() {
		By("deploying Pod with PVC")
		podYAML := `apiVersion: v1
kind: Pod
metadata:
  name: ubuntu
  labels:
	app.kubernetes.io/name: ubuntu
spec:
  containers:
	- name: ubuntu
	  image: quay.io/cybozu/ubuntu:18.04
	  command: ["sleep", "infinity"]
	  volumeDevices:
		- devicePath: /test1
		  name: my-volume
  volumes:
	- name: my-volume
	  persistentVolumeClaim:
		claimName: topo-pvc
`
		claimYAML := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: topo-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: topolvm-provisioner
`
		stdout, stderr, err := ExecAtWithInput(boot0, []byte(claimYAML), "kubectl", "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = ExecAtWithInput(boot0, []byte(podYAML), "kubectl", "apply", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the specified volume exists in the Pod")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "pvc", "topo-pvc", "-n", ns)
			if err != nil {
				return fmt.Errorf("failed to create PVC. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			stdout, stderr, err = ExecAt(boot0, "kubectl", "get", "pods", "ubuntu", "-n", ns)
			if err != nil {
				return fmt.Errorf("failed to create Pod. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "--", "mountpoint", "-d", "/test1")
			if err != nil {
				return fmt.Errorf("failed to check mount point. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "grep", "/test1", "/proc/mounts")
			if err != nil {
				return err
			}
			fields := strings.Fields(string(stdout))
			if fields[2] != "xfs" {
				return errors.New("/test1 is not xfs")
			}
			return nil
		}).Should(Succeed())

		By("writing file under /test1")
		writePath := "/test1/bootstrap.log"
		stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "--", "cp", "/var/log/bootstrap.log", writePath)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "--", "sync")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "--", "cat", writePath)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		Expect(strings.TrimSpace(string(stdout))).ShouldNot(BeEmpty())

		By("getting node name where pod is placed")
		stdout, stderr, err = ExecAt(boot0, "kubectl", "get", "pods", "ubuntu", "-n", ns)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		var pod corev1.Pod
		err = json.Unmarshal(stdout, &pod)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s", stdout)
		nodeName := pod.Spec.NodeName

		By("stopping the node")
		stdout, stderr, err = ExecAt(boot0, "neco", "ipmipower", "stop", nodeName)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		Eventually(func() error {
			return checkNodeStatus(nodeName, "off")
		}).Should(Succeed())

		By("restarting the node")
		stdout, stderr, err = ExecAt(boot0, "neco", "ipmipower", "start", nodeName)
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		Eventually(func() error {
			return checkNodeStatus(nodeName, "on")
		}).Should(Succeed())

		By("confirming that the file exists")
		Eventually(func() error {
			stdout, stderr, err = ExecAt(boot0, "kubectl", "get", "pvc", "topo-pvc", "-n", ns)
			if err != nil {
				return fmt.Errorf("failed to create PVC. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			stdout, stderr, err = ExecAt(boot0, "kubectl", "get", "pods", "ubuntu", "-n", ns)
			if err != nil {
				return fmt.Errorf("failed to create Pod. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", ns, "ubuntu", "--", "cat", writePath)
			if err != nil {
				return fmt.Errorf("failed to cat. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			if len(strings.TrimSpace(string(stdout))) == 0 {
				return fmt.Errorf(writePath + " is empty")
			}
			return nil
		}).Should(Succeed())

		By("getting the volume name")
		stdout, stderr, err = ExecAt(boot0, "kubectl", "get", "pvc", "-n", ns, "topo-pvc", "-o=template", "--template={{.spec.volumeName}}")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		volName := strings.TrimSpace(string(stdout))

		By("deleting the Pod and PVC")
		stdout, stderr, err = ExecAtWithInput(boot0, []byte(podYAML), "kubectl", "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		stdout, stderr, err = ExecAtWithInput(boot0, []byte(claimYAML), "kubectl", "delete", "-n", ns, "-f", "-")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

		By("confirming that the PV is deleted")
		Eventually(func() error {
			stdout, stderr, err = ExecAt(boot0, "kubectl", "get", "pv", volName, "--ignore-not-found")
			if err != nil {
				return fmt.Errorf("failed to get pv/%s. stdout: %s, stderr: %s, err: %v", volName, stdout, stderr, err)
			}
			if len(strings.TrimSpace(string(stdout))) != 0 {
				return fmt.Errorf("target pv exists %s", volName)
			}
			return nil
		}).Should(Succeed())
	})
}

func checkNodeStatus(nodeName, expected string) error {
	stdout, stderr, err := ExecAt(boot0, "neco", "ipmipower", "status", nodeName)
	if err != nil {
		return fmt.Errorf("faild to get node status. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	}
	// The output format is as follows. Get the status string behind ":".
	// "10.72.17.100: on"
	actual := strings.TrimSpace(strings.Split(string(stdout), ":")[1])

	if actual != expected {
		return fmt.Errorf("node status is not yet %s. actual=%s", expected, actual)
	}
	return nil
}
