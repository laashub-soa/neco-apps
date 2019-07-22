package test

import (
	"os"
)

var (
	doBootstrap = os.Getenv("BOOTSTRAP") == "1"
	doUpgrade   = os.Getenv("UPGRADE") == "1"
	boot0       = os.Getenv("BOOT0")
	sshKeyFile  = os.Getenv("SSH_PRIVKEY")
	testID      = os.Getenv("TEST_ID")
	commitID    = os.Getenv("COMMIT_ID")
	externalPID = os.Getenv("EXTERNAL_PID")
)
