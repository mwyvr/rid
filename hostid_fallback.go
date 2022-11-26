//go:build !darwin && !linux && !freebsd && !windows

package rid

import "errors"

func readPlatformMachineID() (string, error) {
	return "", errors.New("not implemented")
}
