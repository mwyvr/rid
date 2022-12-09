package rid

// Verify our custom Base32 character set encoding here:
// https://cryptii.com/pipes/base32

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type idParts struct {
	id        ID
	encoded   string
	encoded64 string
	ts        int64
	rtsig     []byte
	random    uint64
}

var IDs = []idParts{
	// sorted (ascending) should be IDs 2, 3, 0, 1
	{
		// 062ektdeb6039z5masctt333 ts:1670371716697 rtsig:[0x80] random: 58259961960877 | time:2022-12-06 16:08:36.697 -0800 PST ID{0x1,0x84,0xe9,0xe9,0xae,0x59,0x80,0x34,0xfc,0xb4,0x56,0x59,0xad,0xc,0x63}
		ID{0x1, 0x84, 0xe9, 0xe9, 0xae, 0x59, 0x80, 0x34, 0xfc, 0xb4, 0x56, 0x59, 0xad, 0xc, 0x63},
		"062ektdeb6039z5masctt333",
		"AYTp6a5ZgDT8tFZZrQxj",
		1670371716697,
		[]byte{0x80},
		3818124867068038243,
	},
	{
		ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		"zzzzzzzzzzzzzzzzzzzg0000",
		"________________AAAA",
		281474976710655,
		[]byte{0xff},
		18446744073692774400,
	},
	{
		ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		"000000000000000000000000",
		"AAAAAAAAAAAAAAAAAAAA",
		0,
		[]byte{0x00},
		0,
	},
	{
		// 062ektcmm0k3bgwxd4bceqtb ts:1670371710112 rtsig:[0x26] random: 59114275804871 | time:2022-12-06 16:08:30.112 -0800 PST ID{0x1,0x84,0xe9,0xe9,0x94,0xa0,0x26,0x35,0xc3,0x9d,0x69,0x16,0xc7,0x5f,0x4b}
		ID{0x1, 0x84, 0xe9, 0xe9, 0x94, 0xa0, 0x26, 0x35, 0xc3, 0x9d, 0x69, 0x16, 0xc7, 0x5f, 0x4b},
		"062ektcmm0k3bgwxd4bceqtb",
		"AYTp6ZSgJjXDnWkWx19L",
		1670371710112,
		[]byte{0x26},
		3874113179148050251,
	},
}

func TestIDPartsExtraction(t *testing.T) {
	for i, v := range IDs {
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
			if got, want := v.id.Time(), time.UnixMilli(v.ts); got != want {
				t.Errorf("Time() = %v, want %v", got, want)
			}
			if got, want := v.id.Timestamp(), v.ts; got != want {
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
		// Check that timestamp was incremented and is within 30 seconds (30000 ms) of the previous one
		secs := id.Time().Sub(prevID.Time()).Seconds()
		if secs < 0 || secs > 30 {
			t.Error("wrong timestamp in generated ID")
		}
	}
}
func TestIDString(t *testing.T) {
	for _, v := range IDs {
		if got, want := v.encoded, v.id.String(); got != want {
			t.Errorf("String() = %v, want %v", got, want)
		}
	}
}

func TestFromString(t *testing.T) {
	got, err := FromString("062ektcmm0k3bgwxd4bceqtb")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x84, 0xe9, 0xe9, 0x94, 0xa0, 0x26, 0x35, 0xc3, 0x9d, 0x69, 0x16, 0xc7, 0x5f, 0x4b}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
	// nil ID
	got, err = FromString("000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
	// max ID
	got, err = FromString("zzzzzzzzzzzzzzzzzzzzzzzz")
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
}

func TestFromStringInvalid(t *testing.T) {
	_, err := FromString("012345")
	if err != ErrInvalidID {
		t.Errorf("FromString(invalid length) err=%v, want %v", err, ErrInvalidID)
	}
	id, err := FromString("062ez870acdtzd2y3qajilou") // i, l, o, u never in our IDs
	if err != ErrInvalidID {
		t.Errorf("FromString(062ez870acdtzd2y3qajilou - invalid chars) err=%v, want %v", err, ErrInvalidID)
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
			"0000000000000000ilou0000",
			&nilID,
			true,
		},
		{
			"invalid length too long",
			"00000000000000zzzzzzzz00000000000",
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
			"062ektcmm0k3bgwxd4bceqtb",
			&ID{0x1, 0x84, 0xe9, 0xe9, 0x94, 0xa0, 0x26, 0x35, 0xc3, 0x9d, 0x69, 0x16, 0xc7, 0x5f, 0x4b},
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

}

func TestID_UnmarshalTextError(t *testing.T) {
	id := nilID
	if err := id.UnmarshalText([]byte("invalid")); err != ErrInvalidID {
		t.Errorf("ID.UnmarshalText() error = %v, wantErr %v", err, ErrInvalidID)
	}
	id = New() // make a non nil ID
	if err := id.UnmarshalText([]byte("foo")); id != nilID {
		t.Errorf("ID.UnmarshalText() want nil ID, ErrInvalidID, got %v, %v", id, err)
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

type jsonType struct {
	ID  *ID
	Str string
}

func TestIDJSONMarshaling(t *testing.T) {
	// 062ektdeb6039z5masctt333 ts:1670371716697 rtsig:[0x80] random: 58259961960877 | time:2022-12-06 16:08:36.697 -0800 PST ID{0x1,0x84,0xe9,0xe9,0xae,0x59,0x80,0x34,0xfc,0xb4,0x56,0x59,0xad,0xc,0x63}
	id := ID{0x1, 0x84, 0xe9, 0xe9, 0xae, 0x59, 0x80, 0x34, 0xfc, 0xb4, 0x56, 0x59, 0xad, 0xc, 0x63}
	v := jsonType{ID: &id, Str: "test"}
	data, err := json.Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"ID":"062ektdeb6039z5masctt333","Str":"test"}`; got != want {
		t.Errorf("json.Marshal() = %v, want %v", got, want)
	}
}

func TestIDJSONUnmarshaling(t *testing.T) {
	data := []byte(`{"ID":"062ektdeb6039z5masctt333","Str":"test"}`)
	v := jsonType{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x84, 0xe9, 0xe9, 0xae, 0x59, 0x80, 0x34, 0xfc, 0xb4, 0x56, 0x59, 0xad, 0xc, 0x63}
	if got := *v.ID; got.Compare(want) != 0 {
		t.Errorf("json.Unmarshal() = %v, want %v", got, want)
	}
	// should not fail
	err = json.Unmarshal([]byte(`null`), &v)
	if err != nil {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, nil)
	}
}

func TestIDJSONUnmarshalingError(t *testing.T) {
	v := jsonType{}
	// callers are responsible for forcing lower case input for Base32
	// otherwise valid id:
	err := json.Unmarshal([]byte(`{"ID":"062EKTDEB6039Z5MASCTT333"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	// too short
	err = json.Unmarshal([]byte(`{"ID":"062ektdeb6039z5masctt"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	// ID's methods do not know about Base64 encoded data; up to you
	// to use FromString64 as needed
	err = json.Unmarshal([]byte(`{"ID":"AYT3PAVLn1207oRutpGK"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	err = json.Unmarshal([]byte(`{"ID":1}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
}

func TestIDDriverValue(t *testing.T) {
	id := ID{0x1, 0x84, 0xe9, 0xe9, 0x94, 0xa0, 0x26, 0x35, 0xc3, 0x9d, 0x69, 0x16, 0xc7, 0x5f, 0x4b}
	got, err := id.Value()
	if err != nil {
		t.Fatal(err)
	}
	if want := "062ektcmm0k3bgwxd4bceqtb"; got != want {
		t.Errorf("Value() = %v, want %v", got, want)
	}
}

func TestIDDriverScan(t *testing.T) {
	// 062ektcmm0k3bgwxd4bceqtb ts:1670371710112 sig:0x26 rnd: 3874113179148050251 2022-12-06 16:08:30.112 -0800 PST
	// ID{0x1,0x84,0xe9,0xe9,0x94,0xa0,0x26,0x35,0xc3,0x9d,0x69,0x16,0xc7,0x5f,0x4b}
	got := ID{}
	err := got.Scan("062ektcmm0k3bgwxd4bceqtb")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x84, 0xe9, 0xe9, 0x94, 0xa0, 0x26, 0x35, 0xc3, 0x9d, 0x69, 0x16, 0xc7, 0x5f, 0x4b}
	if got.Compare(want) != 0 {
		t.Errorf("Scan() = %v, want %v", got, want)
	}
}

func TestIDDriverScanError(t *testing.T) {
	id := ID{}
	if got, want := id.Scan(0), errors.New("rid: scanning unsupported type: int"); !reflect.DeepEqual(got, want) {
		t.Errorf("Scan() err=%v, want %v", got, want)
	}
	if got, want := id.Scan("0"), ErrInvalidID; got != want {
		t.Errorf("Scan() err=%v, want %v", got, want)
		if id != nilID {
			t.Errorf("Scan() id=%v, want %v", got, nilID)
		}
	}
}

func TestIDDriverScanByteFromDatabase(t *testing.T) {
	got := ID{}
	bs := []byte("062ektcmm0k3bgwxd4bceqtb")
	err := got.Scan(bs)
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x84, 0xe9, 0xe9, 0x94, 0xa0, 0x26, 0x35, 0xc3, 0x9d, 0x69, 0x16, 0xc7, 0x5f, 0x4b}
	if got.Compare(want) != 0 {
		t.Errorf("Scan() = %v, want %v", got, want)
	}
}

var IDList = []ID{IDs[0].id, IDs[1].id, IDs[2].id, IDs[3].id}

func TestSorter_Len(t *testing.T) {
	if got, want := sorter([]ID{}).Len(), 0; got != want {
		t.Errorf("Len() %v, want %v", got, want)
	}
	if got, want := sorter(IDList).Len(), 4; got != want {
		t.Errorf("Len() %v, want %v", got, want)
	}
}

func TestSorter_Less(t *testing.T) {
	// sorted (ascending) should be IDs 2, 3, 0, 1
	sorter := sorter(IDList)
	if !sorter.Less(0, 1) {
		t.Errorf("Less(0, 1) not true")
	}
	if sorter.Less(3, 2) {
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
	// sorted (ascending) should be IDs 2, 3, 0, 1
	if got, want := ids, []ID{IDList[2], IDList[3], IDList[0], IDList[1]}; !reflect.DeepEqual(got, want) {
		t.Fail()
	}
}

func TestString64(t *testing.T) {
	for _, v := range IDs {
		if s64 := String64(v.id); s64 != v.encoded64 {
			t.Errorf("TestString64 %v did not match FromString result %v", s64, v.encoded64)
		}
	}
}

func TestFromString64(t *testing.T) {
	if _, got := FromString64("AYTuxCzTKmROjoji-Ky"); got == nil {
		t.Errorf("Want error %v (too short) got %v", ErrInvalidID, got)
	}
	if _, got := FromString64("AYTuxCzTKmROjoji-KyF00"); got == nil {
		t.Errorf("Want error %v (too long) got %v", ErrInvalidID, got)
	}
	if _, got := FromString64("AYTuxCzTKmROjoji-Ky("); got == nil {
		t.Errorf("Want error %v (invalid char) got %v", ErrInvalidID, got)
	}
	for _, v := range IDs {
		id, err := FromString64(v.encoded64)
		if err != nil {
			t.Error(err)
		}
		b32id, err := FromString(v.encoded)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(id, b32id) {
			t.Errorf("TestFromString64 %v did not match FromString result %v", id, b32id)
		}
	}
}

func Test_runtime_randUint64(t *testing.T) {
	// On a test machine generating 100,000,000 numbers took 24s (4.167 million per second),
	// with zero collisions. 24 seconds = 24,000 milliseconds; 100,000,000/24,000
	// equals ~4,167 unique random generations per millisecond are required.
	// Set at 1,000,000:
	count := 1000000
	exists := make(map[uint64]bool)
	for i := 0; i < count; i++ {
		r := runtime_randUint64()
		if exists[r] {
			t.Errorf("Duplicate random number %d generated within %d attempts", r, count)
		}
		exists[r] = true
	}
}

func Test_runtimeSignature(t *testing.T) {
	// rid.rtsig is set once only and intentionally so in case the platform's
	// machine ID is missing (common on some Linux) and no hostname is returned,
	// not even 'localhost', a random number will be issued. That should be rare.
	//
	// All we are testing is that rtsig should not be a nil value
	var nilMachineID [1]byte
	if got := runtimeSignature(); reflect.DeepEqual(got, nilMachineID) {
		t.Errorf("randomMachineId() = %v, want %v, shouldn't be nil", got, nilMachineID)
	}
	// this should not fail but might fail on CI platforms like Github
	r := runtimeSignature()
	if !reflect.DeepEqual(r, rtsig) {
		t.Errorf("runtimeSignature() = %v does not match pkg init %v.", r, rtsig)
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
	str := "062ez6ecmaky7yksap0arnr6"
	_, err := FromString(str)
	if err != nil {
		b.Error(err)
	}
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
    Timestamp() %d
    RuntimeSignature() %v 
    Random()  %d 
    Time()    %v
    Bytes()   %3v\n`, id.String(), id.Timestamp(), id.RuntimeSignature(), id.Random(), id.Time().UTC(), id.Bytes())
}

func ExampleNewWithTimestamp() {
	id := NewWithTimestamp(uint64(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).UnixMilli()))
	fmt.Printf(`ID: Timestamp() %d Time() %v`, id.Timestamp(), id.Time().UTC())
	// Output: ID: Timestamp() 1577836800000 Time() 2020-01-01 00:00:00 +0000 UTC
}

func ExampleFromString() {
	id, err := FromString("062ez6ecmaky7yksap0arnr6")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Timestamp(), id.Random())
	// Output: 1670467144866 16427575998925264646
}

func ExampleFromString64() {
	id, err := FromString64("AYTvnJruiHfR9aMD96d7")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Timestamp(), id.Random())
	// Output: 1670467328750 8633952041140987771
}
