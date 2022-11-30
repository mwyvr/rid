package rid

import (
	"bytes"
	// "database/sql/driver"
	// enc "encoding"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type idParts struct {
	id        ID
	timestamp int64
	machine   []byte
	pid       uint16
	random    uint32
}

var IDs = []idParts{
	// sorted should be IDs 1, 2, 0
	{
		// [ce0dmp0s249v4q507980] seconds:1669388888 random:1554004572 machine:[0x19, 0x11] pid:5042 time:2022-11-25 07:08:08 -0800 PST
		ID{0x63, 0x80, 0xda, 0x58, 0x19, 0x11, 0x13, 0xb2, 0x5c, 0xa0, 0x3a, 0x50},
		1669388888,
		[]byte{0x19, 0x11},
		5042,
		1554004572,
	},
	{
		ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		0,
		[]byte{0x00, 0x00},
		0x0000,
		0,
	},
	{
		// [ce0djy0s248ra7qrh140] seconds:1669388664 random:519604254 machine:[0x19, 0x11] pid:4485 time:2022-11-25 07:04:24 -0800 PST
		ID{0x63, 0x80, 0xd9, 0x78, 0x19, 0x11, 0x11, 0x85, 0x1e, 0xf8, 0x88, 0x48},
		1669388664,
		[]byte{0x19, 0x11},
		4485,
		519604254,
	},
}

func TestIDPartsExtraction(t *testing.T) {
	for i, v := range IDs {
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
			if got, want := v.id.Time(), time.Unix(v.timestamp, 0); got != want {
				t.Errorf("Time() = %v, want %v", got, want)
			}
			if got, want := v.id.Machine(), v.machine; !bytes.Equal(got, want) {
				t.Errorf("Machine() = %v, want %v", got, want)
			}
			if got, want := v.id.Pid(), v.pid; got != want {
				t.Errorf("Pid() = %v, want %v", got, want)
			}
			if got, want := v.id.Random(), v.random; got != want {
				t.Errorf("Random() = %v, want %v", got, want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	// Generate 10 ids
	var numIDS = 10000
	ids := make([]ID, numIDS)
	for i := 0; i < numIDS; i++ {
		ids[i] = New()
	}
	for i := 1; i < numIDS; i++ {
		prevID := ids[i-1]
		id := ids[i]
		// Test for uniqueness among all other generated ids
		for j, tid := range ids {
			if j != i {
				// can't use ID.Compare for this test, need to compare entire ID[:]
				if bytes.Equal(id[:], tid[:]) {
					t.Errorf("generated ID is not unique (%d/%d)\n%v", i, j, ids)
				}
			}
		}
		// Check that timestamp was incremented and is within 30 seconds of the previous one
		secs := id.Time().Sub(prevID.Time()).Seconds()
		if secs < 0 || secs > 30 {
			t.Error("wrong timestamp in generated ID")
		}
		// Check that machine ids are the same
		if !bytes.Equal(id.Machine(), prevID.Machine()) {
			t.Error("machine ID not equal")
		}
		// Check that pids are the same
		if id.Pid() != prevID.Pid() {
			t.Error("pid not equal")
		}
	}
}

func TestIDString(t *testing.T) {
	id := ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}
	if got, want := id.String(), "9p4e2pv0yj3e8a215q4g"; got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestIDEncode(t *testing.T) {
	id := ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}
	text := make([]byte, encodedLen)
	if got, want := string(id.Encode(text)), "9p4e2pv0yj3e8a215q4g"; got != want {
		t.Errorf("Encode() = %v, want %v", got, want)
	}
}

func TestFromString(t *testing.T) {
	got, err := FromString("9p4e2pv0yj3e8a215q4g")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
}

func TestFromStringInvalid(t *testing.T) {
	_, err := FromString("invalid")
	if err != ErrInvalidID {
		t.Errorf("FromString(invalid) err=%v, want %v", err, ErrInvalidID)
	}
	id, err := FromString("ce0cnw0s25j1ksgsilou") // i, l, o, u never in our IDs
	if err != ErrInvalidID {
		t.Errorf("FromString(ce0cnw0s25j1ksgsilou - invalid chars) err=%v, want %v", err, ErrInvalidID)
	}
	if id != nilID {
		t.Errorf("FromString() =%v, there want %v", id, nilID)
	}
}

func TestID_IsNil(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want bool
	}{
		{
			name: "ID not nil",
			id:   New(),
			want: false,
		},
		{
			name: "Nil ID",
			id:   ID{},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got, want := tt.id.IsNil(), tt.want; got != want {
				t.Errorf("IsNil() = %v, want %v", got, want)
			}
		})
	}
}

func TestNilID(t *testing.T) {
	got := ID{}
	if want := NilID(); !reflect.DeepEqual(got, want) {
		t.Error("NilID() not equal ID{}")
	}
}

func TestNilID_IsNil(t *testing.T) {
	if !NilID().IsNil() {
		t.Error("NilID().IsNil() is not true")
	}
}

func TestFromBytes_Invariant(t *testing.T) {
	want := New()
	got, err := FromBytes(want.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if got.Compare(want) != 0 {
		t.Error("FromBytes(id.Bytes()) != id")
	}
}

func TestFromBytes_InvalidBytes(t *testing.T) {
	cases := []struct {
		length     int
		shouldFail bool
	}{
		{11, true},
		{12, false},
		{13, true},
	}
	for _, c := range cases {
		b := make([]byte, c.length)
		_, err := FromBytes(b)
		if got, want := err != nil, c.shouldFail; got != want {
			t.Errorf("FromBytes() error got %v, want %v", got, want)
		}
	}
}

func TestIDDriverScan(t *testing.T) {

	// [ce0djy0s248ra7qrh140] seconds:1669388664 random:519604254 machine:[0x19, 0x11] pid:4485 time:2022-11-25 07:04:24 -0800 PST
	// ID{0x63, 0x80, 0xd9, 0x78, 0x19, 0x11, 0x11, 0x85, 0x1e, 0xf8, 0x88, 0x48}
	got := ID{}
	err := got.Scan("ce0djy0s248ra7qrh140")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x63, 0x80, 0xd9, 0x78, 0x19, 0x11, 0x11, 0x85, 0x1e, 0xf8, 0x88, 0x48}

	if got.Compare(want) != 0 {
		t.Errorf("Scan() = %v, want %v", got, want)
	}
}

var IDList = []ID{IDs[0].id, IDs[1].id, IDs[2].id}

func TestSorter_Len(t *testing.T) {
	if got, want := sorter([]ID{}).Len(), 0; got != want {
		t.Errorf("Len() %v, want %v", got, want)
	}
	if got, want := sorter(IDList).Len(), 3; got != want {
		t.Errorf("Len() %v, want %v", got, want)
	}
}

func TestSorter_Less(t *testing.T) {
	sorter := sorter(IDList)
	if !sorter.Less(1, 0) {
		t.Errorf("Less(1, 0) not true")
	}
	if sorter.Less(2, 1) {
		t.Errorf("Less(2, 1) true")
	}
	if sorter.Less(0, 0) {
		t.Errorf("Less(0, 0) true")
	}
}

func TestSorter_Swap(t *testing.T) {
	ids := make([]ID, 0)
	ids = append(ids, IDList...)
	sorter := sorter(ids)
	sorter.Swap(0, 1)
	if got, want := ids[0], IDList[1]; !reflect.DeepEqual(got, want) {
		t.Error("ids[0] != IDList[1]")
	}
	if got, want := ids[1], IDList[0]; !reflect.DeepEqual(got, want) {
		t.Error("ids[1] != IDList[0]")
	}
	sorter.Swap(2, 2)
	if got, want := ids[2], IDList[2]; !reflect.DeepEqual(got, want) {
		t.Error("ids[2], IDList[2]")
	}
}

func TestSort(t *testing.T) {
	ids := make([]ID, 0)
	ids = append(ids, IDList...)
	Sort(ids)
	if got, want := ids, []ID{IDList[1], IDList[2], IDList[0]}; !reflect.DeepEqual(got, want) {
		t.Fail()
	}
}

// Benchmarks
func BenchmarkNew(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = New()
		}
	})
}

// common use case
func BenchmarkNewString(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = New().String()
		}
	})
}

// encoding performance
func BenchmarkString(b *testing.B) {
	id := New()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = id.String()
		}
	})
}

// decoding performance
func BenchmarkFromString(b *testing.B) {
	str := "ce1tcars24hcmnsc8jvg"
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = FromString(str)
		}
	})
}

// examples
func ExampleNew() {
	id := New()
	fmt.Printf(`ID:
    String()  %s   
    Seconds() %d  
    Random()  %d // random for this one-off run 
    Time()    %v
    Bytes()   %3v\n`, id.String(), id.Seconds(), id.Random(), id.Time(), id.Bytes())
}

func ExampleNewWithTimestamp() {
	id := NewWithTimestamp(uint32(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()))
	fmt.Printf(`ID:
    String()  %s
    Seconds() %d
    Machine() %v 
    Pid()     %d
    Random()  %d 
    Time()    %v
    Bytes()   %3v\n`, id.String(), id.Seconds(), id.Machine(), id.Pid(), id.Random(), id.Time().UTC(), id.Bytes())
}

func ExampleFromString() {
	id, err := FromString("ce0dz5gs24h2e30a74rg")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Seconds(), id.Random())
	// 1669390230 201996556
}
