// +build kind

package test

import (
	"bytes"
	"net"
	"os/exec"
	"strings"

	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

var (
	sshClients = make(map[string]*sshAgent)
)

type sshAgent struct {
	client *ssh.Client
	conn   net.Conn
}

func sshTo(address string, sshKey ssh.Signer, userName string) (*sshAgent, error) {
	return nil, nil
}

func prepareSSHClients(addresses ...string) error {
	return nil
}

// ExecAt executes command at given host
func ExecAt(host string, args ...string) (stdout, stderr []byte, e error) {
	return ExecAtWithInput(host, nil, args...)
}

// ExecAtWithInput executes command at given host with input
// WARNING: `input` can contain secret data.  Never output `input` to console.
func ExecAtWithInput(host string, input []byte, args ...string) (stdout, stderr []byte, e error) {
	return doExec(nil, input, args...)
}

func doExec(agent *sshAgent, input []byte, args ...string) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("sh", "-c", strings.Join(args, " "))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}

	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

// ExecSafeAt executes command at given host and returns only stdout
func ExecSafeAt(host string, args ...string) []byte {
	stdout, stderr, err := ExecAt(host, args...)
	ExpectWithOffset(1, err).To(Succeed(), "[%s] %v, stdout: %s, stderr: %s", host, args, stdout, stderr)
	return stdout
}
