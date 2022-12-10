//go:build linux

package rid

import "os"

// see https://0pointer.de/blog/projects/ids.html
func readPlatformMachineID() (string, error) {
	b, err := os.ReadFile("/var/lib/dbus/machine-id")
	if err != nil || len(b) == 0 {
		b, err = os.ReadFile("/proc/sys/kernel/random/boot_id")
	}
	return string(b), err
}
