package test

import (
	"errors"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testTeleport() {
	It("should deploy teleport node services", func() {
		By("retrieving LoadBalancer IP address of teleport auth service")
		var addr string
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "-n", "teleport", "get", "service", "teleport-auth",
				"--output=jsonpath={.status.loadBalancer.ingress[0].ip}")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			ret := strings.TrimSpace(string(stdout))
			if len(ret) == 0 {
				return errors.New("teleport auth IP address is empty")
			}
			addr = ret
			return nil
		}).Should(Succeed())

		By("storing LoadBalancer IP address to etcd")
		ExecSafeAt(boot0, "env", "ETCDCTL_API=3", "etcdctl", "--cert=/etc/etcd/backup.crt", "--key=/etc/etcd/backup.key",
			"put", "/neco/teleport/auth-servers", `[\"`+addr+`:3025\"]`)

		By("starting teleport node services on boot servers")
		for _, h := range []string{boot0, boot1, boot2} {
			ExecSafeAt(h, "sudo", "neco", "teleport", "config")
			ExecSafeAt(h, "sudo", "systemctl", "start", "teleport-node.service")
		}
	})
}
