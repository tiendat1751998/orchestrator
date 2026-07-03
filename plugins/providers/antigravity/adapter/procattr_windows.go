//go:build windows

package adapter

import (
	"fmt"
	"os/exec"
	"syscall"
)

// newProcAttr returns platform-specific process group attributes.
// On Windows, it configures the process to run in a new process group.
func newProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// killProcessGroup kills the target process and all children in its process group.
// On Windows, it invokes taskkill with /F (force) and /T (tree/group) flags to cleanly terminate the process tree.
// If the process has already exited (exit code 128), it returns nil.
func killProcessGroup(pid int) error {
	// ponytail: taskkill is standard on windows for killing process groups
	cmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid))
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 128 indicates the process was not found (already exited)
			if exitErr.ExitCode() == 128 {
				return nil
			}
		}
		return err
	}
	return nil
}
