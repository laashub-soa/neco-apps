package test

import (
	. "github.com/onsi/ginkgo"
)

// TestSetup tests "neco setup"
func testSetup() {
	It("should get nodes", func() {
		execSafeAt(boot0, "ckecli", "kubernetes", "issue", ">", ".kube/config")
		execSafeAt(boot0, "kubectl", "get", "nodes")
	})
}
