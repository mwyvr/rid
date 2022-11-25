package rid

// TODO add chronological sorting test

import (
	"bytes"
	// "database/sql/driver"
	// enc "encoding"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

var (
	// testing concurrency safety
	wg            sync.WaitGroup
	numConcurrent = 2000 // go routines X
	numIter       = 500  // id creation/routine
)

type idParts struct {
    id ID 
    timestamp int64
    machine []byte
    pid     uint16
    random  uint32
}

var IDs = []idParts {
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
	ids := make([]ID, 100)
	for i := 0; i < 100; i++ {
		ids[i] = New()
	}
	for i := 1; i < 100; i++ {
		prevID := ids[i-1]
		id := ids[i]
		// Test for uniqueness among all other 9 generated ids
		for j, tid := range ids {
			if j != i {
                // can't use ID.Compare for this test, need to compare entire ID[:]
                if bytes.Compare(id[:], tid[:]) == 0 {

				// if id.Compare(tid) == 0 {
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
func BenchmarkIDNew(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next(){
            _ = New()
        }
    })
}

func BenchmarkIDNewEncoded(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next(){
            _ = New().String()
        }
    })
}

// examples
func ExampleNew() {
	id := New()
	fmt.Printf(`ID:
    String()       %s   
    Seconds() %d  
    Random()        %d // random for this one-off run 
    Time()         %v
    Bytes():       %3v  
`, id.String(), id.Seconds(), id.Random(), id.Time(), id.Bytes())
}

func ExampleNewWithTime() {
	id := NewWithTime(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
	fmt.Printf(`ID:
    String()       %s
    Seconds() %d
    Random()        %d // random for this one-off run 
    Time()         %v
    Bytes():       %3v
`, id.String(), id.Seconds(), id.Random(), id.Time().UTC(), id.Bytes())
}

func ExampleFromString() {
	id, err := FromString("05yx13hj9kq4g")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Seconds(), id.Random())
	// [05yx13hj9kq4g] ms:1639881519692 count:61000 time:2021-12-18 18:38:39.692 -0800 PST id:{1, 125, 208, 142, 50, 76, 238, 72}
}

// func TestID_MarshalJSON(t *testing.T) {
// 	if got, err := nilID.MarshalJSON(); string(got) != "null" {
// 		t.Errorf("ID.MarshalJSON() of nilID error = %v, got %v", err, got)
// 	}
// 	if got, err := (ID{1, 125, 208, 142, 50, 76, 238, 72}).MarshalJSON(); string(got) != "\"05yx13hj9kq4g\"" {
// 		if err != nil {
// 			t.Errorf("ID.MarshalJSON() err %v marshaling %v", err, "\"05yx13hj9kq4g\"")
// 		}
// 		t.Errorf("ID.MarshalJSON() got %v want %v", string(got), "\"05yx13hj9kq4g\"")
// 	}
// }
//
// func TestID_UnmarshalJSON(t *testing.T) {
// 	var id ID
// 	err := id.UnmarshalJSON([]byte("null"))
// 	if err != nil {
// 		t.Errorf("ID.UnmarshalJSON() of null, error = %v", err)
// 	}
// 	if id != nilID {
// 		t.Errorf("ID.UnmarshalJSON() error = %v", err)
// 	}
// 	// 2020...
// 	text := []byte("\"05yykgvzqc002\"")
// 	if err = id.UnmarshalJSON(text); err != nil {
// 		t.Errorf("ID.UnmarshalJSON() error = %v", err)
//
// 	} else if id != (ID{1, 125, 233, 195, 127, 187, 0, 1}) {
// 		t.Errorf("ID.UnmarshalJSON() of %v, got %v", text, id.String())
// 	}
// }
