package ingress

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

var _ = BeforeSuite(func() {
	test.RunBeforeSuite()
	test.ExecAt(test.Boot0, "kubectl", "create", "namespace", "test-ingress")
})

var _ = AfterSuite(func() {
	test.ExecAt(test.Boot0, "kubectl", "delete", "namespace", "test-ingress")
})

// This must be the only top-level test container.
// Other tests and test containers must be listed in this.
var _ = Describe("GitOps Test", func() {
	Context("contour", testContour)
})
