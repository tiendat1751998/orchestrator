//go:build !windows

package adapter

import (
	"syscall"
)

// newProcAttr returns platform-specific process group attributes.
// On Unix, it configures the process to set its process group ID (PGID).
func newProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// killProcessGroup kills the target process and all children in its process group.
// On Unix, sending a signal to a negative PID kills the entire process group.
// If the process has already exited (syscall.ESRCH), it returns nil.
func killProcessGroup(pid int) error {
	// ponytail: syscall.Kill to negative pid kills the entire process group
	err := syscall.Kill(-pid, syscall.SIGKILL)
	if err == syscall.ESRCH {
		return nil
	}
	return err
}
