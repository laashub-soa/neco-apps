package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
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

	It("should access to boot servers via teleport ssh", func() {
		By("getting the address of teleport-proxy")
		stdout, stderr, err := ExecAt(boot0, "kubectl", "-n", "teleport", "get", "service", "teleport-proxy",
			"--output=jsonpath={.status.loadBalancer.ingress[0].ip}")
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		addr := string(stdout)

		By("create user")
		stdout, stderr, err = ExecAt(boot0, "kubectl", "-n", "teleport", "exec", "teleport-auth-0", "tctl", "users", "add", "cybozu", "cybozu,root")
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		slashSplited := strings.Split(strings.Split(string(stdout), "\n")[1], "/")
		token := slashSplited[len(slashSplited)-1]
		payload, err := json.Marshal(map[string]string{
			"invite_token":        token,
			"pass":                "dummypassword",
			"second_factor_token": "",
			"user":                "cybozu",
		})
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		cmd := exec.Command("curl", "--insecure", "-H", "\"Content-Type: application/json; charset=UTF-8\"", "-d", "'"+string(payload)+"'", "https://"+addr+":30080/v1/webapi/users")
		err = cmd.Run()
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)

		By("accessing boot servers using tsh command")
		// for _, n := range []string{"boot-0", "boot-1", "boot-2"} {
		// 	cmd := exec.Command("tsh", "--proxy="+addr+":3080", "--user=cybozu", "cybozu@"+n, "date")
		// 	err = cmd.Run()
		// 	Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		// }
	})
}
