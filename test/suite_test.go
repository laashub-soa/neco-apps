package test

import (
	"os"
	"testing"

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

var _ = BeforeSuite(RunBeforeSuite)

// This must be the only top-level test container.
// Other tests and test containers must be listed in this.
var _ = Describe("GitOps Test", func() {
	Context("setup", testSetup)
	Context("metallb", testMetalLB)
	Context("external-dns", testExternalDNS)
	//Context("cert-manager", testCertManager)
	Context("contour", testContour)
	Context("machines-endpoints", testMachinesEndpoints)
	Context("kube-state-metrics", testKubeStateMetrics)
	Context("prometheus", testPrometheus)
	Context("alertmanager", testAlertmanager)
	Context("metrics", testMetrics)
})
