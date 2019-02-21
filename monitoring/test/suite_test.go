package monitoring

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

	RegisterFailHandler(Fail)
	RunSpecs(t, "Test")
}

var _ = BeforeSuite(test.RunBeforeSuite)

// This must be the only top-level test container.
// Other tests and test containers must be listed in this.
var _ = Describe("GitOps Test", func() {
	Context("monitoring", testMonitoring)
	Context("machines-endpoints", testMachinesEndpoints)
})
