package test

import (
	"os"
	"testing"

	"github.com/cybozu-go/neco-ops/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	if os.Getenv("SSH_PRIVKEY") == "" {
		t.Skip("no SSH_PRIVKEY envvar")
	}

	_, err := os.Stat("../../test/account.json")
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("Google Service Account file does not exist.  Skip Cloud DNS-related tests.")
		}
		t.Fatal(err)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Test")
}

var _ = BeforeSuite(func() {
	test.RunBeforeSuite()
})

// This must be the only top-level test container.
// Other tests and test containers must be listed in this.
var _ = Describe("GitOps Test", func() {
	Context("Setup", testSetup)
	Context("External DNS", testExternalDNS)
})
