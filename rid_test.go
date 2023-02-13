// Acknowledgement: This source file is based on work in package github.com/rs/xid,
// a zero-configuration globally-unique ID generator. See LICENSE.rs-xid.
package rid

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
	id      ID
	encoded string
	ts      int64
	random  uint64
}

var IDs = []idParts{
	// sorted (ascending) should be IDs 2, 3, 0, 5, 4, 1
	{
		// dfp7emzzzzy30ey2 ts:1672246995 rnd:281474912761794 2022-12-28 09:03:15 -0800 PST ID{0x63,0xac,0x76,0xd3,0xff,0xff,0xfc,0x30,0x37,0xc2}
		ID{0x63, 0xac, 0x76, 0xd3, 0xff, 0xff, 0xfc, 0x30, 0x37, 0xc2},
		"dfp7emzzzzy30ey2",
		1672246995,
		281474912761794,
	},
	{
		// zzzzzzzzzzzzzzzz ts:4294967295 rnd:281474976710655 2106-02-06 22:28:15 -0800 PST ID{0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff}
		ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		"zzzzzzzzzzzzzzzz",
		4294967295,
		281474976710655,
	},
	{
		// 0000000000000000 ts:0 rnd:              0 1969-12-31 16:00:00 -0800 PST ID{0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0}
		ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		"0000000000000000",
		0,
		0,
	},
	{
		// dfp7em00001p0t5j ts:1672246992 rnd:       56649906 2022-12-28 09:03:12 -0800 PST ID{0x63,0xac,0x76,0xd0,0x0,0x0,0x3,0x60,0x68,0xb2}
		ID{0x63, 0xac, 0x76, 0xd0, 0x0, 0x0, 0x3, 0x60, 0x68, 0xb2},
		"dfp7em00001p0t5j",
		1672246992,
		56649906,
	},
	{
		// dgb58zr000000000 ts:1674859647 rnd:              0 2023-01-27 14:47:27 -0800 PST ID{0x63,0xd4,0x54,0x7f,0x0,0x0,0x0,0x0,0x0,0x0}
		ID{0x63, 0xd4, 0x54, 0x7f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		"dgb58zr000000000",
		1674859647,
		0,
	},
	{
		// dgb53lewel4ndk94 ts:1674858957 rnd:207175420364068 2023-01-27 14:35:57 -0800 PST ID{0x63,0xd4,0x51,0xcd,0xbc,0x6c,0xc9,0x56,0x45,0x24}
		ID{0x63, 0xd4, 0x51, 0xcd, 0xbc, 0x6c, 0xc9, 0x56, 0x45, 0x24},
		"dgb53lewel4ndk94",
		1674858957,
		207175420364068,
	},
}

func TestIDPartsExtraction(t *testing.T) {
	for i, v := range IDs {
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
			if got, want := v.id.Time(), time.Unix(v.ts, 0); got != want {
				t.Errorf("Time() = %v, want %v", got, want)
			}
			if got, want := v.id.Timestamp(), v.ts; got != want {
				t.Errorf("Time() = %v, want %v", got, want)
			}
			if got, want := v.id.Random(), v.random; got != want {
				t.Errorf("Random() = %v, want %v", got, want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	// Generate N ids, see if all unique
	// Parallel generation test is in ./cmd/eval/uniqcheck/main.go
	numIDS := 1000
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
				// can't use ID.Compare for this test as it compares only the time
				// component of IDs
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

func TestIDEncode(t *testing.T) {
	id := ID{0x63, 0xac, 0x76, 0xd3, 0xff, 0xff, 0xfc, 0x30, 0x37, 0xc2}
	text := make([]byte, encodedLen)
	if got, want := string(id.Encode(text)), "dfp7emzzzzy30ey2"; got != want {
		t.Errorf("Encode() = %v, want %v", got, want)
	}
}

func TestFromString(t *testing.T) {
	// dfp7emzzzzy30ey2 ts:1672246995 rnd:281474912761794 2022-12-28 09:03:15 -0800 PST ID{0x63,0xac,0x76,0xd3,0xff,0xff,0xfc,0x30,0x37,0xc2}
	got, err := FromString("dfp7emzzzzy30ey2")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x63, 0xac, 0x76, 0xd3, 0xff, 0xff, 0xfc, 0x30, 0x37, 0xc2}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
	// nil ID
	got, err = FromString("0000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
	// max ID
	got, err = FromString("zzzzzzzzzzzzzzzz")
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
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
			// 0000000000000000 ts:0 rnd:              0 1969-12-31 16:00:00 -0800 PST ID{0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0}
			"valid",
			"0000000000000000",
			&nilID,
			false,
		},
		{
			"invalid chars",
			"000000000000000u",
			&nilID,
			true,
		},
		{
			"invalid length too long",
			"12345678901",
			&nilID,
			true,
		},
		{
			"invalid length too short",
			"dfb7emm",
			&nilID,
			true,
		},
		{
			// dfp7emzzzzy30ey2 ts:1672246995 rnd:281474912761794 2022-12-28 09:03:15 -0800 PST ID{0x63,0xac,0x76,0xd3,0xff,0xff,0xfc,0x30,0x37,0xc2}
			"valid id",
			"dfp7emzzzzy30ey2",
			&ID{0x63, 0xac, 0x76, 0xd3, 0xff, 0xff, 0xfc, 0x30, 0x37, 0xc2},
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
	if bytes.Compare(got[:], want[:]) != 0 {
		t.Error("FromBytes(id.Bytes()) != id")
	}
	// invalid
	got, err = FromBytes([]byte{0x1, 0x2})
	if bytes.Compare(got[:], nilID[:]) != 0 {
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
	// dfp7emzzzzy30ey2 ts:1672246995 rnd:281474912761794 2022-12-28 09:03:15 -0800 PST ID{0x63,0xac,0x76,0xd3,0xff,0xff,0xfc,0x30,0x37,0xc2}
	id := ID{0x63, 0xac, 0x76, 0xd3, 0xff, 0xff, 0xfc, 0x30, 0x37, 0xc2}
	v := jsonType{ID: &id, Str: "test"}
	data, err := json.Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"ID":"dfp7emzzzzy30ey2","Str":"test"}`; got != want {
		t.Errorf("json.Marshal() = %v, want %v", got, want)
	}
}

func TestIDJSONUnmarshaling(t *testing.T) {
	// dfp7emzzzzy30ey2 ts:1672246995 rnd:281474912761794 2022-12-28 09:03:15 -0800 PST ID{0x63,0xac,0x76,0xd3,0xff,0xff,0xfc,0x30,0x37,0xc2}
	data := []byte(`{"ID":"dfp7emzzzzy30ey2","Str":"test"}`)
	v := jsonType{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x63, 0xac, 0x76, 0xd3, 0xff, 0xff, 0xfc, 0x30, 0x37, 0xc2}
	if got := *v.ID; bytes.Compare(got[:], want[:]) != 0 {
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
	err := json.Unmarshal([]byte(`{"ID":"DFP8T54NN0JZ37HW"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	// too short
	err = json.Unmarshal([]byte(`{"ID":"dfp8t54nn0jz37h"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	// no 'a' in character set
	err = json.Unmarshal([]byte(`{"ID":"0000000000000a"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	// invalid on multiple levels
	err = json.Unmarshal([]byte(`{"ID":1}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
}

func TestIDDriverValue(t *testing.T) {
	// dfp7emzzzzy30ey2 ts:1672246995 rnd:281474912761794 2022-12-28 09:03:15 -0800 PST ID{0x63,0xac,0x76,0xd3,0xff,0xff,0xfc,0x30,0x37,0xc2}
	id := ID{0x63, 0xac, 0x76, 0xd3, 0xff, 0xff, 0xfc, 0x30, 0x37, 0xc2}
	got, err := id.Value()
	if err != nil {
		t.Fatal(err)
	}
	if want := "dfp7emzzzzy30ey2"; got != want {
		t.Errorf("Value() = %v, want %v", got, want)
	}
}

func TestIDDriverScan(t *testing.T) {
	// dfp7emzzzzy30ey2 ts:1672246995 rnd:281474912761794 2022-12-28 09:03:15 -0800 PST ID{0x63,0xac,0x76,0xd3,0xff,0xff,0xfc,0x30,0x37,0xc2}
	got := ID{}
	err := got.Scan("dfp7emzzzzy30ey2")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x63, 0xac, 0x76, 0xd3, 0xff, 0xff, 0xfc, 0x30, 0x37, 0xc2}
	if bytes.Compare(got[:], want[:]) != 0 {
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
	// dfp7emzzzzy30ey2 ts:1672246995 rnd:281474912761794 2022-12-28 09:03:15 -0800 PST ID{0x63,0xac,0x76,0xd3,0xff,0xff,0xfc,0x30,0x37,0xc2}
	got := ID{}
	bs := []byte("dfp7emzzzzy30ey2")
	err := got.Scan(bs)
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x63, 0xac, 0x76, 0xd3, 0xff, 0xff, 0xfc, 0x30, 0x37, 0xc2}
	if bytes.Compare(got[:], want[:]) != 0 {
		t.Errorf("Scan() = %v, want %v", got, want)
	}
}

func TestFromBytes_InvalidBytes(t *testing.T) {
	cases := []struct {
		length     int
		shouldFail bool
	}{
		{rawLen - 1, true},
		{rawLen, false},
		{rawLen + 1, true},
	}
	for _, c := range cases {
		b := make([]byte, c.length)
		_, err := FromBytes(b)
		if got, want := err != nil, c.shouldFail; got != want {
			t.Errorf("FromBytes() error got %v, want %v", got, want)
		}
	}
}

func TestCompare(t *testing.T) {
	pairs := []struct {
		left     ID
		right    ID
		expected int
	}{
		{IDs[1].id, IDs[0].id, 1},
		{ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, IDs[2].id, 0},
		{IDs[0].id, IDs[0].id, 0},
		{IDs[2].id, IDs[1].id, -1},
		{IDs[5].id, IDs[4].id, -1},
	}
	for _, p := range pairs {
		if p.expected != p.left.Compare(p.right) {
			t.Errorf("%s Compare to %s should return %d", p.left, p.right, p.expected)
		}
		if -1*p.expected != p.right.Compare(p.left) {
			t.Errorf("%s Compare to %s should return %d", p.right, p.left, -1*p.expected)
		}
	}
}

var IDList = []ID{IDs[0].id, IDs[1].id, IDs[2].id, IDs[3].id, IDs[4].id, IDs[5].id}

func TestSorter_Len(t *testing.T) {
	if got, want := sorter([]ID{}).Len(), 0; got != want {
		t.Errorf("Len() %v, want %v", got, want)
	}
	if got, want := sorter(IDList).Len(), 6; got != want {
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
	// sorted (ascending) should be IDs 2, 3, 0, 5, 4, 1
	if got, want := ids, []ID{IDList[2], IDList[3], IDList[0], IDList[5], IDList[4], IDList[1]}; !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot %v\nwant %v\n", got, want)
	}
}

func TestFastrand48(t *testing.T) {
	// see eval/uniqcheck/main.go for additional testing
	t.Run("check-dupes", func(t *testing.T) {
		var id [rawLen]byte
		keys := make(map[[rawLen]byte]bool) // keys can be arrays, not slices
		count := 5000000
		for i := 0; i < count; i++ {
			r := New()
			copy(id[:], r[:])
			if keys[id] {
				t.Errorf("Duplicate random number %d generated within %d attempts", id, count)
			}
			keys[id] = true
		}
	})
	t.Run("check-bounds", func(t *testing.T) {
		count := 10000000
		for i := 0; i < count; i++ {
			r := fastrand48()
			if r > maxRandom {
				t.Errorf("Random number %d exceeds maxRandom %d", r, maxRandom)
			}
		}
	})
}

// Benchmarks
// globals & func locals added to avoid compiler over-optimization and silly results
var (
	benchResultID     ID
	benchResultString string
)

// Create new ID
func BenchmarkNew(b *testing.B) {
	var r ID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = New()
		}
		benchResultID = r
	})
}

// common use case, generate an ID, encode as a string:
func BenchmarkNewString(b *testing.B) {
	var r string
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = New().String()
		}
		benchResultString = r
	})
}

// encoding performance only
func BenchmarkString(b *testing.B) {
	id := New()
	var r string
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = id.String()
		}
		benchResultString = r
	})
}

// decoding performance only
func BenchmarkFromString(b *testing.B) {
	var r ID
	// dfp7emzzzzy30ey2 ts:1672246995 rnd:281474912761794 2022-12-28 09:03:15 -0800 PST ID{0x63,0xac,0x76,0xd3,0xff,0xff,0xfc,0x30,0x37,0xc2}
	str := "dfp7emzzzzy30ey2"
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, _ = FromString(str)
		}
		benchResultID = r
	})
}

// examples
func ExampleNew() {
	id := New()
	fmt.Printf(`ID:
    String()  %s
    Timestamp() %d
    Random()  %d 
    Time()    %v
    Bytes()   %3v\n`, id.String(), id.Timestamp(), id.Random(), id.Time().UTC(), id.Bytes())
}

func ExampleNewWithTime() {
	id := NewWithTime(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
	fmt.Printf(`ID: Timestamp() %d Time() %v`, id.Timestamp(), id.Time().UTC())
	// Output: ID: Timestamp() 1577836800 Time() 2020-01-01 00:00:00 +0000 UTC
}

func ExampleFromString() {
	// dfp7emzzzzy30ey2 ts:1672246995 rnd:281474912761794 2022-12-28 09:03:15 -0800 PST ID{0x63,0xac,0x76,0xd3,0xff,0xff,0xfc,0x30,0x37,0xc2}
	id, err := FromString("dfp7emzzzzy30ey2")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Timestamp(), id.Random())
	// Output: 1672246995 281474912761794
}
