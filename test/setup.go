package test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TestSetup tests "neco setup"
func TestSetup() {
	It("should get nodes", func() {
		execSafeAt(boot0, "kubectl", "get", "nodes")
	})
}
