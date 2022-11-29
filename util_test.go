package rid

import (
	"reflect"
	"testing"
)

func Test_randomMachineId(t *testing.T) {
	// highly unlikely although possible to match
	mID := randomMachineId()
	if got := randomMachineId(); reflect.DeepEqual(got, mID) {
		t.Errorf("randomMachineId() = %v, want %v, shouldn't be equal", got, mID)
	}
}

func Test_readMachineID(t *testing.T) {
	// should not be a nil value
	var nilMachineID = make([]byte, 2)
	if got := randomMachineId(); reflect.DeepEqual(got, nilMachineID) {
		t.Errorf("randomMachineId() = %v, want %v, shouldn't be nil", got, nilMachineID)
	}
}
