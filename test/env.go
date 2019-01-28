package test

import (
	"os"
)

const (
	ArgoCDNamespace = "argocd"
)

var (
	Boot0      = os.Getenv("BOOT0")
	SSHKeyFile = os.Getenv("SSH_PRIVKEY")
	TestID     = os.Getenv("TEST_ID")
	CommitID   = os.Getenv("COMMIT_ID")
)
