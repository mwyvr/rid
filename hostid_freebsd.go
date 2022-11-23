// +build freebsd

package rid

import "syscall"

func readPlatformMachineID() (string, error) {
	return syscall.Sysctl("kern.hostuuid")
}
