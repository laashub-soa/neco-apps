package test

import (
	"os"
)

var (
	boot0         = os.Getenv("BOOT0")
    sshKeyFile    = os.Getenv("SSH_PRIVKEY")
)
