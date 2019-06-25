package test

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cybozu-go/sabakan/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
)

func testNetworkPolicy() {
	It("should create test-netpol namespace", func() {
		ExecSafeAt(boot0, "kubectl", "delete", "namespace", "test-netpol", "--ignore-not-found=true")
		ExecSafeAt(boot0, "kubectl", "create", "namespace", "test-netpol")
	})

	It("should accept and deny packets according to the registered network policies", func() {

		By("deployment Pods")
		deployYAML := `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: testhttpd
  namespace: test-netpol
spec:
  replicas: 2
  selector:
    matchLabels:
      run: testhttpd
  template:
    metadata:
      labels:
        run: testhttpd
    spec:
      containers:
      - image: quay.io/cybozu/testhttpd:0
        name: testhttpd
      restartPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  name: testhttpd
  namespace: test-netpol
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8000
  selector:
    run: testhttpd
---
apiVersion: crd.projectcalico.org/v1
kind: NetworkPolicy
metadata:
  name: ingress-httpdtest
  namespace: test-netpol
spec:
  order: 1000.0
  selector: run == 'testhttpd'
  types:
    - Ingress
  ingress:
    - action: Allow
      protocol: TCP
      destination:
        ports:
          - 8000
`
		_, stderr, err := ExecAtWithInput(boot0, []byte(deployYAML), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		By("deploy ubuntu for network commands")
		debugYAML := `
apiVersion: v1
kind: Pod
metadata:
  name: ubuntu
  labels:
    app: ubuntu
spec:
  securityContext:
    runAsUser: 10000
    runAsGroup: 10000
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu-debug:18.04
    command: ["sleep", "infinity"]`
		_, stderr, err = ExecAtWithInput(boot0, []byte(debugYAML), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		By("checking hostname is resolved by cluster-dns")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "exec", "ubuntu", "--", "nslookup", "-timeout=10", "testhttpd.test-netpol")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		By("checking hostname out of cluster can be resolved")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "exec", "ubuntu", "--", "nslookup", "-timeout=10", "cybozu.com")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			return nil
		}).Should(Succeed())

		By("getting pod list")
		stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "pods", "-A", "-o=json")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

		podList := new(corev1.PodList)
		err = json.Unmarshal(stdout, podList)
		Expect(err).NotTo(HaveOccurred())

		stdout, stderr, err = ExecAt(boot0, "kubectl", "get", "pods", "-n", "test-netpol", "-o=json")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

		testhttpdPodList := new(corev1.PodList)
		err = json.Unmarshal(stdout, testhttpdPodList)
		Expect(err).NotTo(HaveOccurred())

		By("checking ping is dropped")
		for _, pod := range podList.Items {
			if pod.Spec.HostNetwork {
				continue
			}
			By(fmt.Sprintf("  -> ping to %s[%s]", pod.GetName(), pod.Status.PodIP))
			stdout, stderr, err := ExecAt(boot0, "ping", "-c", "1", "-W", "2", pod.Status.PodIP)
			Expect(err).To(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
		}

		By("checking connection")
		const portShouldBeDenied = 65535

		testcase := []struct {
			namespace string
			selector  string
			ports     []int
		}{
			{"argocd", "app.kubernetes.io/name=argocd-application-controller", []int{8082}},
			{"argocd", "app.kubernetes.io/name=argocd-redis", []int{6379}},
			{"argocd", "app.kubernetes.io/name=argocd-repo-server", []int{8081, 8084}},
			{"argocd", "app.kubernetes.io/name=argocd-server", []int{8080, 8083}},
			{"external-dns", "app=cert-manager", []int{9402}},
			{"external-dns", "app=webhook", []int{6443}},
			{"external-dns", "app.kubernetes.io/name=external-dns", []int{7979}},
			{"ingress", "app=contour", []int{8002, 8080, 8443}},
			{"internet-egress", "k8s-app=squid", []int{3128}},
			{"internet-egress", "k8s-app=unbound", []int{53}},
			{"kube-system", "cke.cybozu.com/appname=cluster-dns", []int{1053, 8080}},
			{"kube-system", "k8s-app=kube-state-metrics", []int{8080, 8081}},
			{"metallb-system", "component=controller", []int{7472}},
			{"monitoring", "app=alertmanager", []int{9093}},
			{"monitoring", "app=prometheus", []int{9090}},
		}

		for _, tc := range testcase {
			By(fmt.Sprintf("getting pod list: ns=%s, selector=%s", tc.namespace, tc.selector))
			stdout, stderr, err := ExecAt(boot0, "kubectl", "-n", tc.namespace, "-l", tc.selector, "get", "pods", "-o=json")
			Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

			podList := new(corev1.PodList)
			err = json.Unmarshal(stdout, podList)
			Expect(err).NotTo(HaveOccurred())

			for _, pod := range podList.Items {
				By(fmt.Sprintf("checking pod: %s[%s]", pod.GetName(), pod.Status.PodIP))
				for _, port := range tc.ports {
					By(fmt.Sprintf("  -> allowed port: %d", port))
					stdout, stderr, err = ExecAtWithInput(boot0, []byte("Xclose"), "timeout", "3s", "telnet", pod.Status.PodIP, strconv.Itoa(port), "-e", "X")
					switch t := err.(type) {
					case *ssh.ExitError:
						// telnet command returns 124 when it times out
						Expect(t.ExitStatus()).NotTo(Equal(124))
					default:
						Expect(err).NotTo(HaveOccurred())
					}
				}

				By(fmt.Sprintf("  -> denied port: %d", portShouldBeDenied))
				stdout, stderr, err = ExecAtWithInput(boot0, []byte("Xclose"), "timeout", "3s", "telnet", pod.Status.PodIP, strconv.Itoa(portShouldBeDenied), "-e", "X")
				switch t := err.(type) {
				case *ssh.ExitError:
					// telnet command returns 124 when it times out
					Expect(t.ExitStatus()).To(Equal(124))
				default:
					Expect(err).NotTo(HaveOccurred())
				}

				if tc.namespace == "internet-egress" {
					By("access to local IP")
					testhttpdIP := testhttpdPodList.Items[0].Status.PodIP
					stdout, stderr, err = ExecAt(boot0, "kubectl", "exec", "-n", pod.Namespace, pod.Name, "--", "curl", testhttpdIP, "-m", "5")
					Expect(err).To(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
				}
			}
		}

		By("checking ping to the idrac subnet is dropped by network policy")
		stdout, stderr, err = ExecAt(boot0, "sabactl", "machines", "get")
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

		var machines []sabakan.Machine
		err = json.Unmarshal(stdout, &machines)
		Expect(err).ShouldNot(HaveOccurred())
		for _, m := range machines {
			By("  -> ping to " + m.Spec.BMC.IPv4)
			stdout, _, err := ExecAt(boot0, "kubectl", "exec", "ubuntu", "--", "ping", "-W", "3", "-c", "1", m.Spec.BMC.IPv4)
			Expect(err).To(HaveOccurred(), "stdout: %s", stdout)
		}
	})
}
