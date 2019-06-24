package test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cybozu-go/sabakan/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sparrc/go-ping"
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
			By(fmt.Sprintf("sending ping to pod: %s[%s]", pod.GetName(), pod.Status.PodIP))
			pinger, err := ping.NewPinger(pod.Status.PodIP)
			if err != nil {
				Expect(err).NotTo(HaveOccurred())
			}

			pinger.Timeout = 2 * time.Second
			pinger.Count = 1
			pinger.SetPrivileged(true)
			pinger.Run()
			stats := pinger.Statistics()
			Expect(stats.PacketsRecv).To(Equal(0))
		}

		By("checking connection")
		const portShouldBeDenied = 65535

		testcase := []struct {
			podNamePrefix  string
			ports          []int
			internetEgress bool
		}{
			{"argocd-application-controller", []int{8082}, false},
			{"argocd-redis", []int{6379}, false},
			{"argocd-repo-server", []int{8081, 8084}, false},
			{"argocd-server", []int{8080, 8083}, false},
			{"cert-manager", []int{9402}, false},
			{"external-dns", []int{7979}, false},
			{"contour", []int{8002, 8080, 8443}, false},
			{"squid", []int{53, 3128}, true},
			{"unbound", []int{53}, true},
			{"cluster-dns", []int{1053, 8080}, false},
			{"coil-node", []int{9383}, false},
			{"kube-state-metrics ", []int{8080, 8081}, false},
			{"controller", []int{7472}, false},
			{"speaker", []int{7472}, false},
			{"alertmanager", []int{9093}, false},
			{"prometheus", []int{9090}, false},
		}

		for _, tc := range testcase {
			By("getting pod list: " + tc.podNamePrefix)
			var targetPods []corev1.Pod
			for _, pod := range podList.Items {
				if strings.HasPrefix(pod.GetName(), tc.podNamePrefix) {
					targetPods = append(targetPods, pod)
				}
			}
			Expect(len(targetPods)).NotTo(Equal(0), "pod is not found: %s", tc.podNamePrefix)

			for _, pod := range targetPods {
				By(fmt.Sprintf("checking pod: %s[%s]", pod.GetName(), pod.Status.PodIP))
				for _, port := range tc.ports {
					By(fmt.Sprintf("dialing to allowed port: %d", port))
					stdout, stderr, err = ExecAtWithInput(boot0, []byte("Xclose"), "timeout", "3s", "telnet", pod.Status.PodIP, strconv.Itoa(port), "-e", "X")
					Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
				}

				By(fmt.Sprintf("dialing to denied port: %d", portShouldBeDenied))
				stdout, stderr, err = ExecAtWithInput(boot0, []byte("Xclose"), "timeout", "3s", "telnet", pod.Status.PodIP, strconv.Itoa(portShouldBeDenied), "-e", "X")
				switch t := err.(type) {
				case *ssh.ExitError:
					// telnet command returns 124 when it times out
					Expect(t.ExitStatus).To(Equal(124))
				default:
					Expect(err).NotTo(HaveOccurred())
				}

				if tc.internetEgress {
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
			By("sending: " + m.Spec.BMC.IPv4)
			stdout, _, err := ExecAt(boot0, "kubectl", "exec", "ubuntu", "--", "ping", "-W", "3", "-c", "1", m.Spec.BMC.IPv4)
			Expect(err).To(HaveOccurred(), "stdout: %s", stdout)
		}
	})
}
