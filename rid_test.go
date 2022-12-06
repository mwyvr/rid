package rid

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type idParts struct {
	id     ID
	ts     int64
	rtsig  []byte
	random uint64
}

var IDs = []idParts{
	// sorted (ascending) should be IDs 1, 2, 0
	{
		// ce6s9m4nv5be5w91b2tg seconds:1670223056 rtsig:[0x95,0xd9] random: 95532708092085 | time:2022-12-04 22:50:56 -0800 PST ID{0x63,0x8d,0x94,0xd0,0x95,0xd9,0x56,0xe2,0xf1,0x21,0x58,0xb5}
		ID{0x63, 0x8d, 0x94, 0xd0, 0x95, 0xd9, 0x56, 0xe2, 0xf1, 0x21, 0x58, 0xb5},
		1670223056,
		[]byte{0x95, 0xd9},
		95532708092085,
	},
	{
		ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		0,
		[]byte{0x00, 0x00},
		0,
	},
	{
		// ce6s7tarwkqzbch94xt0 seconds:1670222825 rtsig:[0x58,0xe4] random:263838535067508 | time:2022-12-04 22:47:05 -0800 PST ID{0x63,0x8d,0x93,0xe9,0x58,0xe4,0xef,0xf5,0xb2,0x29,0x27,0x74}
		ID{0x63, 0x8d, 0x93, 0xe9, 0x58, 0xe4, 0xef, 0xf5, 0xb2, 0x29, 0x27, 0x74},
		1670222825,
		[]byte{0x58, 0xe4},
		263838535067508,
	},
}

func TestIDPartsExtraction(t *testing.T) {
	for i, v := range IDs {
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
			if got, want := v.id.Time(), time.Unix(v.ts, 0); got != want {
				t.Errorf("Time() = %v, want %v", got, want)
			}
			if got, want := v.id.RuntimeSignature(), v.rtsig; !bytes.Equal(got, want) {
				t.Errorf("RuntimeSignature() = %v, want %v", got, want)
			}
			if got, want := v.id.Random(), v.random; got != want {
				t.Errorf("Random() = %v, want %v", got, want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	// Generate N ids, see if all unique
	// TODO add parallel test
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
		if !bytes.Equal(id.RuntimeSignature(), prevID.RuntimeSignature()) {
			t.Error("machine ID not equal")
		}
	}
}

func TestIDString(t *testing.T) {
	id := ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}
	if got, want := id.String(), "9p4e2pv0yj3e8a215q4g"; got != want {
		t.Errorf("String() = %v, want %v", got, want)
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
	// nil ID
	got, err = FromString("00000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
	// max ID
	got, err = FromString("zzzzzzzzzzzzzzzzzzzg") // trailing z also equals due to padding
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
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

func TestID_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		id      *ID
		wantErr bool
	}{
		{
			"invalid chars",
			"0000000000000000ilou",
			&nilID,
			true,
		},
		{
			"invalid length too long",
			"0000000000000000000000000",
			&nilID,
			true,
		},
		{
			"invalid length too short",
			"abcde",
			&nilID,
			true,
		},
		{
			"valid id",
			"ce7m7d3evfg1pa4t93x0",
			&ID{0x63, 0x8f, 0x43, 0xb4, 0x6e, 0xdb, 0xe0, 0x1b, 0x28, 0x9a, 0x48, 0xfa},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.id.UnmarshalText([]byte(tt.text)); (err != nil) != tt.wantErr {
				t.Errorf("ID.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	id := New()
	if err := id.UnmarshalText([]byte("foo")); err != ErrInvalidID {
		t.Errorf("ID.UnmarshalText() error = %v, wantErr %v", err, ErrInvalidID)
	}
	if err := id.UnmarshalText([]byte("foo")); err != nil && !id.IsNil() {
		t.Errorf("ID.UnmarshalText() error = %v, want nil ID, got %v", err, id)
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

// is an alias function, no need to repeat above, just for coverage report
func TestID_IsZero(t *testing.T) {
	id := ID{}
	if !id.IsZero() {
		t.Errorf("ID.IsZero() = %v, want %v", id.IsZero(), true)
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
	// invalid
	got, err = FromBytes([]byte{0x1, 0x2})
	if got.Compare(nilID) != 0 {
		t.Error("FromBytes([]byte{0x1, 0x2}) - invalid - != nilID")
	}
	if err == nil {
		t.Fatal(err)
	}
}

func TestIDDriverScan(t *testing.T) {

	// ce7kerav4yj65mzy1rp0 seconds:1670330209 rtsig:[0x5b,0x27] random:180744370392620 | time:2022-12-06 04:36:49 -0800 PST
	// ID{0x63,0x8f,0x37,0x61,0x5b,0x27,0xa4,0x62,0xd3,0xfe,0xe,0x2c}
	got := ID{}
	err := got.Scan("ce7kerav4yj65mzy1rp0")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x63, 0x8f, 0x37, 0x61, 0x5b, 0x27, 0xa4, 0x62, 0xd3, 0xfe, 0xe, 0x2c}
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

func Test_randUint32(t *testing.T) {
	one := randUint32()
	another := randUint32()
	if one == another {
		t.Errorf("randUint32() = %v should not equal another: %v", one, another)
	}
}

func Test_randUint64(t *testing.T) {
	one := randUint64()
	another := randUint64()
	if one == another {
		t.Errorf("randUint64() = %v should not equal another: %v", one, another)
	}
}

func Test_runtimeSignature(t *testing.T) {
	// should not be a nil value
	var nilMachineID = make([]byte, 2)
	if got := runtimeSignature(); reflect.DeepEqual(got, nilMachineID) {
		t.Errorf("randomMachineId() = %v, want %v, shouldn't be nil", got, nilMachineID)
	}
}

// Benchmarks
// Create new ID
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
    RuntimeSignature() %v 
    Random()  %d 
    Time()    %v
    Bytes()   %3v\n`, id.String(), id.Seconds(), id.RuntimeSignature(), id.Random(), id.Time().UTC(), id.Bytes())
}

func ExampleNewWithTimestamp() {
	id := NewWithTimestamp(uint32(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()))
	fmt.Printf(`ID:
    String()  %s
    Seconds() %d
    RuntimeSignature() %v 
    Random()  %d 
    Time()    %v
    Bytes()   %3v\n`, id.String(), id.Seconds(), id.RuntimeSignature(), id.Random(), id.Time().UTC(), id.Bytes())
}

func ExampleFromString() {
	id, err := FromString("ce7k5pagzd2kvnrbtt60")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Seconds(), id.Random())
	// 1670329049 76131903198860
}
