package test

import (
	"encoding/json"
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ping "github.com/sparrc/go-ping"
	corev1 "k8s.io/api/core/v1"
)

func testNetworkPolicy() {
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
    - Egress
  ingress:
    - action: Allow
      protocol: TCP
      destination:
        ports:
          - 8000
  egress:
    - action: Allow
`
		_, stderr, err := ExecAtWithInput(boot0, []byte(deployYAML), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		By("checking hostname is resolved by cluster-dns")
		ips, err := net.LookupIP("testhttpd.test-netpol")
		Expect(len(ips)).To(Equal(2))

		By("checking ping is dropped")
		stdout, stderr, err := ExecAt(boot0, "kubectl", "--namespace=kube-system", "get", "pods", "--selector=k8s-app=coil-controllers", "-o=json")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		podList := new(corev1.PodList)
		err = json.Unmarshal(stdout, podList)
		Expect(err).NotTo(HaveOccurred())

		for _, pod := range podList.Items {
			pinger, err := ping.NewPinger(pod.Status.PodIP)
			if err != nil {
				Expect(err).NotTo(HaveOccurred())
			}

			pinger.Count = 1
			pinger.SetPrivileged(true)
			pinger.Run()
			stats := pinger.Statistics()
			Expect(stats.PacketsRecv).To(Equal(0))
		}

		By("checking connect to open ports")
		testcase := []struct {
			prefix string
			ports  []int
		}{
			{"argocd-application-controller", []int{8082}},
			{"argocd-redis", []int{6379}},
			{"argocd-repo-server", []int{8081, 8084}},
			{"argocd-server", []int{8080, 8083}},
			{"cert-manager", []int{9402}},
			{"external-dns", []int{7979}},
			{"contour", []int{8002, 8080, 8443}},
			{"squid", []int{53, 3128}},
			{"unbound", []int{53}},
			{"cluster-dns", []int{1053, 8080}},
			{"coil-node", []int{9383}},
			{"kube-state-metrics ", []int{8080, 8081}},
			{"controller", []int{7472}},
			{"speaker", []int{7472}},
			{"alertmanager", []int{9093}},
			{"prometheus", []int{9090}},
    }
    
    for _, tc := range testcase {
      for _, pod := range podList.Items {
      }
    }
	})
}
