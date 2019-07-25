package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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
		By("adding proxy addr entry to /etc/hosts")
		stdout, stderr, err := ExecAt(boot0, "kubectl", "-n", "teleport", "get", "service", "teleport-proxy",
			"--output=jsonpath={.status.loadBalancer.ingress[0].ip}")
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		addr := string(stdout)
		f, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		_, err = f.Write([]byte("\n" + addr + " teleport.gcp0.dev-ne.co\n"))
		f.Close()

		By("creating user")
		stdout, stderr, err = ExecAt(boot0, "kubectl", "-n", "teleport", "exec", "teleport-auth-0", "tctl", "users", "add", "cybozu", "cybozu,root")
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		slashSplited := strings.Split(strings.Split(string(stdout), "\n")[1], "/")
		token := slashSplited[len(slashSplited)-1]
		payload, err := json.Marshal(map[string]string{
			"invite_token":        token,
			"pass":                "dummypass",
			"second_factor_token": "",
			"user":                "cybozu",
		})
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		cmd := exec.Command("curl", "--fail", "--insecure", "-H", "Content-Type: application/json; charset=UTF-8", "-d", string(payload), "https://teleport.gcp0.dev-ne.co:3080/v1/webapi/users")
		err = cmd.Run()
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)

		By("logging in using tsh command")
		cmd = exec.Command("tsh", "--insecure", "--proxy=teleport.gcp0.dev-ne.co:3080", "--user=cybozu", "login")
		stdin, err := cmd.StdinPipe()
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		err = cmd.Start()
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		_, err = io.WriteString(stdin, "dummypass\n")
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		err = cmd.Wait()
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)

		By("accessing boot servers using tsh command")
		for _, n := range []string{"boot-0", "boot-1", "boot-2"} {
			cmd := exec.Command("tsh", "--insecure", "--proxy=teleport.gcp0.dev-ne.co:3080", "--user=cybozu", "cybozu@"+n, "date")
			err = cmd.Run()
			Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)
		}
	})
}
