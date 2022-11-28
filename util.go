package rid

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

var (
	rander = rand.Reader
)

// randomBytes completely fills slice b with random data via crypto/rand
func randomBytes(b []byte) {
	// as this should *never* fail, panic is appropriate
	if _, err := io.ReadFull(rander, b); err != nil {
		panic(fmt.Errorf("rid: cannot generate random number: %v;", err))
	}
}

// randomMachineId generates a fallback machine ID
func randomMachineId() []byte {
	b := make([]byte, 2)
	randomBytes(b)
	return b
}

// readMachineId generates machine id and puts it into the machineId global
// variable. If this function fails to get the hostname, and the fallback
// fails, it will cause a runtime error.
func readMachineID() []byte {
	id := make([]byte, 2)
	hid, err := readPlatformMachineID()
	if err != nil || len(hid) == 0 {
		hid, err = os.Hostname()
	}
	if err == nil && len(hid) != 0 {
		hw := md5.New()
		hw.Write([]byte(hid))
		copy(id, hw.Sum(nil))
	} else {
		// Fallback to rand number if machine id can't be gathered
		id = randomMachineId()
	}
	return id
}
