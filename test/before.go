package test

import (
	"fmt"
	"time"

	"github.com/cybozu-go/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// RunBeforeSuite is for Ginkgo BeforeSuite.
func RunBeforeSuite() {
	fmt.Println("Preparing...")

	SetDefaultEventuallyPollingInterval(time.Second)
	SetDefaultEventuallyTimeout(10 * time.Minute)

	err := prepareSSHClients(boot0)
	Expect(err).NotTo(HaveOccurred())

	// sync VM root filesystem to store newly generated SSH host keys.
	for h := range sshClients {
		ExecSafeAt(h, "sync")
	}

	log.DefaultLogger().SetOutput(GinkgoWriter)

	fmt.Println("Begin tests...")
}
